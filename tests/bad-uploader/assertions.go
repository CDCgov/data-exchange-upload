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

func NewCheck(ctx context.Context, conf *config, c TestCase, uploadUrl string) *UploadCheck {
	httpClient := &http.Client{}
	if conf.tokenSource != nil {
		httpClient = oauth2.NewClient(ctx, conf.tokenSource)
	}
	_, uploadId := path.Split(uploadUrl)

	var eventChecker EventChecker
	if reportsURL != "" {
		eventChecker = &PSAPIEventChecker{
			GraphQLClient: graphql.NewClient(reportsURL, nil),
		}
	}

	return &UploadCheck{
		Case:        c,
		UploadId:    uploadId,
		InfoClient:  httpClient,
		EventClient: eventChecker,
	}
}

type UploadCheck struct {
	Case        TestCase
	UploadId    string
	InfoClient  *http.Client
	EventClient EventChecker
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

func CheckDelivery(ctx context.Context, check *UploadCheck) error {
	err := withRetry(ctx, check.CheckInfo)
	if err != nil {
		return err
	}
	slog.Info("verified upload", "upload", check.UploadId)

	return nil
}

func CheckEvents(ctx context.Context, check *UploadCheck) error {
	if check.EventClient == nil {
		slog.Info("no event source provided; skipping event check")
		return nil
	}
	var reports []Report
	err := withRetry(ctx, func() error {
		var err error
		reports, err = check.EventClient.Events(ctx, check.UploadId)
		if err != nil {
			return errors.Join(err, &ErrFatalAssertion{"failed to fetch events"})
		}

		err = compareReports(reports, check.Case.ExpectedReports)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	// If the file doesn't exist, create it, or append to the file
	// TODO drop in output dir
	// TODO only output if failed
	f, err := os.OpenFile(path.Base(check.UploadId)+".reports", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	je := json.NewEncoder(f)
	if err := je.Encode(reports); err != nil {
		return err
	}
	return nil
}

func withRetry(timeout context.Context, checker CheckFunc) error {
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			// Perform the checkable action
			err := checker()
			if err == nil {
				// Action was successful, all done
				return nil
			}
			// Did we get a fatal error?
			var fatalErr *ErrFatalAssertion
			if errors.As(err, &fatalErr) {
				return err
			}
			slog.Debug("non-fatal error during retryable check: ", "error", err)
		case <-timeout.Done():
			return fmt.Errorf("failed to perform check in time")
		}
	}
}

func compareReports(actual []Report, expected []Report) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("not the right number of reports %d %d", len(actual), len(expected))
	}

	var errs error
	for i, e := range expected {
		if actual[i].StageInfo.Action != e.StageInfo.Action {
			errs = errors.Join(errs, fmt.Errorf("expected report missing: index %d, expected %s, actual %s", i, e.StageInfo.Action, actual[i].StageInfo.Action))
		}
	}
	return errs
}
