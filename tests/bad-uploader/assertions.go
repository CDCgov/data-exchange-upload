package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"sort"
	"time"

	"github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

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

func Check(ctx context.Context, c TestCase, upload string, conf *config) error {
	serverUrl, _ := path.Split(url)
	infoUrl, err := neturl.JoinPath(serverUrl, "info", path.Base(upload))
	if err != nil {
		return err
	}
	httpClient := &http.Client{}
	if conf.tokenSource != nil {
		httpClient = oauth2.NewClient(ctx, conf.tokenSource)
	}
	resp, err := httpClient.Get(infoUrl)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to check upload: %d, %s", resp.StatusCode, infoUrl)
	}
	slog.Info("verified upload", "upload", infoUrl)

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
					"id": path.Base(upload),
				}

				if err := client.Query(context.Background(), &q, variables); err != nil {
					return err
				}
				reports := q.GetReports
				sort.Sort(reports)
				slog.Debug("Expecting", "reports", c.ExpectedReports)

				errs = errors.Join(errs, compareReports(reports, c.ExpectedReports))
				// If the file doesn't exist, create it, or append to the file
				f, err := os.OpenFile(path.Base(upload)+".reports", os.O_CREATE|os.O_WRONLY, 0644)
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
