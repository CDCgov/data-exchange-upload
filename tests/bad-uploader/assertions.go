package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"slices"
	"sort"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

type ErrFatalAssertion struct {
	m string
}

func (e *ErrFatalAssertion) Error() string {
	return e.m
}

type Report struct {
	ID        string
	UploadId  string
	StageName string
	Timestamp time.Time
	Content   json.RawMessage
}

type Reports []Report

func (r Reports) Len() int           { return len(r) }
func (r Reports) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r Reports) Less(i, j int) bool { return r[i].Timestamp.Before(r[j].Timestamp) }

func NewCheck(ctx context.Context, conf *config, c TestCase, uploadUrl string) *UploadCheck {
	httpClient := &http.Client{}
	if conf.tokenSource != nil {
		httpClient = oauth2.NewClient(ctx, conf.tokenSource)
	}
	_, uploadId := path.Split(uploadUrl)
	return &UploadCheck{
		Case:       c,
		UploadId:   uploadId,
		InfoClient: httpClient,
	}
}

type UploadCheck struct {
	Case       TestCase
	UploadId   string
	InfoClient *http.Client
}

func (uc *UploadCheck) CheckInfo() error {
	serverUrl, _ := path.Split(url)
	infoUrl, err := neturl.JoinPath(serverUrl, "info", path.Base(uc.UploadId))
	if err != nil {
		return err
	}

	resp, err := uc.InfoClient.Get(infoUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to check upload info: %d, %s", resp.StatusCode, infoUrl)
	}

	var info InfoResponse
	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &info)
	if err != nil {
		return err
	}

	// Check info fields
	if info.UploadStatus.Status != "Complete" {
		return fmt.Errorf("upload unsuccessful for upload ID %s: %v", uc.UploadId, info.UploadStatus)
	}

	// Check delivery targets
	if len(info.Deliveries) != len(uc.Case.ExpectedDeliveryTargets) {
		return fmt.Errorf("expected %d deliveries but got %d", len(uc.Case.ExpectedDeliveryTargets), len(info.Deliveries))
	}

	for _, delivery := range info.Deliveries {
		if delivery.Status != "SUCCESS" {
			return &ErrFatalAssertion{
				m: fmt.Sprintf("%s delivery failed: %v", delivery.Name, delivery.Issues),
			}
		}

		if !slices.Contains(uc.Case.ExpectedDeliveryTargets, delivery.Name) {
			return &ErrFatalAssertion{
				m: fmt.Sprintf("delivery target should be one of %v but got %s", uc.Case.ExpectedDeliveryTargets, delivery.Name),
			}
		}
	}

	return nil
}

type CheckFunc func() error

func Check(ctx context.Context, check *UploadCheck) error {
	/*
		Always check info endpoint
		  - Upload status
		  - Delivery status
		    - Needs to be based on target within config
		Check events
		  - Use interface for this.  File vs API
		  - Always check file.  Check API if URL provided
		  - TODO: config flag to skip this step as can be brittle
	*/
	err := withRetry(ctx, check.CheckInfo)
	if err != nil {
		return err
	}
	slog.Info("verified upload", "upload", check.UploadId)

	if reportsURL != "" {
		timer := time.NewTicker(1 * time.Second)
		for {
			var errs error
			select {
			case <-timer.C:

				client := graphql.NewClient(reportsURL, nil)

				var q struct {
					GetReports Reports `graphql:"getReports(uploadId: $id, reportsSortedBy: null, sortOrder: null)"`
				}

				variables := map[string]interface{}{
					"id": path.Base(check.UploadId),
				}

				if err := client.Query(context.Background(), &q, variables); err != nil {
					return err
				}
				reports := q.GetReports
				sort.Sort(reports)
				slog.Debug("Expecting", "reports", check.Case.ExpectedReports)

				errs = errors.Join(errs, compareReports(reports, check.Case.ExpectedReports))
				// If the file doesn't exist, create it, or append to the file
				f, err := os.OpenFile(path.Base(check.UploadId)+".reports", os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
				defer f.Close()
				je := json.NewEncoder(f)
				if err := je.Encode(reports); err != nil {
					return err
				}

				if errs == nil {
					slog.Debug("validated run", "reports", reports)
					return nil
				}
			case <-ctx.Done():
				return errors.Join(errs, fmt.Errorf("failed to validate upload in time"))
			}
		}
	}

	return nil
}

func withRetry(timeout context.Context, checker CheckFunc) error {
	for {
		// Perform the checkable action
		err := checker()
		if err == nil {
			// Action was successful, all done
			return nil
		}
		// Did we timeout yet?
		if errors.Is(timeout.Err(), context.Canceled) {
			// Yes, return and notify caller
			return timeout.Err()
		}
		var fatalErr *ErrFatalAssertion
		if errors.As(err, &fatalErr) {
			return err
		}
		// No, wait and perform check action again
		time.Sleep(1 * time.Second)
	}
}

func compareReports(actual []Report, expected []Report) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("not the right number of reports %d %d", len(actual), len(expected))
	}

	var errs error
	for i, e := range expected {
		if actual[i].StageName != e.StageName {
			errs = errors.Join(errs, fmt.Errorf("expected report missing: index %d, expected %s", i, e))
		}
	}
	return errs
}
