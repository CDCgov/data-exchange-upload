package testing

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/tus/tusd/v2/pkg/handler"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	ts          *httptest.Server
	testContext context.Context
)

func TestTus(t *testing.T) {
	url := ts.URL

	for name, c := range Cases {
		tuid, err := RunTusTestCase(url, "test/test.txt", c)
		time.Sleep(2 * time.Second) // Hard delay to wait for all non-blocking hooks to finish.

		if err != nil {
			t.Error(name, err)
		} else {
			if tuid != "" {
				config, err := metadata.GetConfigFromManifest(testContext, handler.MetaData(c.metadata))
				if err != nil {
					t.Fatal(err)
				}

				expectedMetadataTransformReportCount := 2
				if v, ok := c.metadata["version"]; !ok || v == "1.0" {
					expectedMetadataTransformReportCount = 3
				}

				reportSummary, err := readReportFile(tuid)
				if err != nil {
					t.Error("failed to read report file for", "tuid", tuid)
				}
				err = checkReportSummary(reportSummary, reports.StageMetadataVerify, 1)
				if err != nil {
					t.Error(err.Error())
				}
				err = checkReportSummary(reportSummary, reports.StageMetadataTransform, expectedMetadataTransformReportCount)
				if err != nil {
					t.Error(err.Error())
				}
				err = checkReportSummary(reportSummary, reports.StageUploadStatus, 1)
				if err != nil {
					t.Error(err.Error())
				}
				err = checkReportSummary(reportSummary, reports.StageUploadStarted, 1)
				if err != nil {
					t.Error(err.Error())
				}
				err = checkReportSummary(reportSummary, reports.StageUploadCompleted, 1)
				if err != nil {
					t.Error(err.Error())
				}
				err = checkReportSummary(reportSummary, reports.StageFileCopy, 3)
				if err != nil {
					t.Error(err.Error())
				}

				if c.err != nil {
					metadataVerifyReport, ok := reportSummary.Summaries[reports.StageMetadataVerify]
					if !ok {
						t.Error("expected metadata verify report but got none")
					}
					if metadataVerifyReport.Reports[0].StageInfo.Issues == nil {
						t.Error("expected reported issues but got none", name, tuid, metadataVerifyReport.Reports[0])
					}

					uploadStatusReport, ok := reportSummary.Summaries[reports.StageUploadStatus]
					if !ok {
						t.Error("expected upload status report but got none")
					}
					if uploadStatusReport.Reports[0].Content.(reports.UploadStatusContent).Offset != uploadStatusReport.Reports[0].Content.(reports.UploadStatusContent).Size {
						t.Error("expected latest status report to have equal offset and size but were different", name, tuid, uploadStatusReport.Reports[0])
					}
				}

				// Post-processing
				events, err := readEventFile(tuid)
				if err != nil {
					t.Error("no events found for tuid", "tuid", tuid)
				}
				if len(events) != len(config.Copy.Targets) {
					t.Errorf("expected %d file ready event(s) but got %d", len(config.Copy.Targets), len(events))
				}

				// Also check that the .meta file exists in the dex folder.
				if _, err := os.Stat("./test/uploads/" + tuid + ".meta"); errors.Is(err, os.ErrNotExist) {
					t.Error("meta file was not copied to dex checkpoint for file", tuid)
				}

				if slices.Contains(config.Copy.Targets, "edav") {
					// Check that the file exists in the target checkpoint folders.
					if _, err := os.Stat("./test/edav/" + tuid); errors.Is(err, os.ErrNotExist) {
						t.Error("file was not copied to edav checkpoint for file", tuid)
					}
				}

				if slices.Contains(config.Copy.Targets, "routing") {
					if _, err := os.Stat("./test/routing/" + tuid); errors.Is(err, os.ErrNotExist) {
						t.Error("file was not copied to routing checkpoint for file", tuid)
					}
				}

				// Also check that the metadata in the .meta file is hydrated with v2 manifest fields.
				metaFile, err := os.Open("./test/uploads/" + tuid + ".meta")
				if err != nil {
					t.Error("error opening meta file for file", tuid)
				}
				defer metaFile.Close()

				bytes, _ := io.ReadAll(metaFile)
				var processedMeta map[string]string
				err = json.Unmarshal(bytes, &processedMeta)
				if err != nil {
					t.Error("error deserializing metadata for file", tuid)
				}

				translationFields := map[string]string{
					"meta_destination_id": "data_stream_id",
					"meta_ext_event":      "data_stream_route",
				}
				v, ok := c.metadata["version"]

				if !ok || v == "1.0" {
					for v1Key, v2Key := range translationFields {
						v1Val, ok := processedMeta[v1Key]
						if !ok {
							t.Error("malformed metadata; missing required field", v1Key, processedMeta)
						}
						v2Val, ok := processedMeta[v2Key]
						if !ok {
							t.Error("v1 metadata not hydrated; missing v2 field", v2Key)
						}
						if v1Val != v2Val {
							t.Error("v1 to v2 fields not properly translated", v1Val, v2Val)
						}
					}
				}
			}

			t.Log("test case", name, "passed", tuid)
		}
	}
}

func TestWellKnownEndpoints(t *testing.T) {
	endpoints := []string{
		"/",
		"/health",
		"/version",
	}
	client := ts.Client()
	for _, endpoint := range endpoints {
		resp, err := client.Get(ts.URL + endpoint)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			t.Error("bad response for ", endpoint, resp.StatusCode)
		}
	}
}

func TestGetFileDeliveryPrefixDate(t *testing.T) {
	ctx := context.TODO()
	m := map[string]string{
		"version":           "2.0",
		"data_stream_id":    "test_stream",
		"data_stream_route": "test_route",
	}
	metadata.Cache.SetConfig("v2/test_stream-test_route.json", &validation.ManifestConfig{
		Copy: validation.CopyConfig{
			FolderStructure: metadata.FolderStructureDate,
		},
	})

	p, err := metadata.GetFilenamePrefix(ctx, m)
	if err != nil {
		t.Fatal(err)
	}
	dateTokens := strings.Split(p, "/")
	if len(dateTokens) != 4 {
		t.Error("date prefix not properly formatted", p)
	}
}

func TestGetFileDeliveryPrefixRoot(t *testing.T) {
	ctx := context.TODO()
	m := map[string]string{
		"version":           "2.0",
		"data_stream_id":    "test_stream",
		"data_stream_route": "test_route",
	}
	metadata.Cache.SetConfig("v2/test_stream-test_route.json", &validation.ManifestConfig{
		Copy: validation.CopyConfig{
			FolderStructure: metadata.FolderStructureRoot,
		},
	})

	p, err := metadata.GetFilenamePrefix(ctx, m)
	if err != nil {
		t.Fatal(err)
	}
	if p != "" {
		t.Error("expected file delivery prefix to be empty but was", p)
	}
}

func TestDeliveryFilenameSuffixUploadId(t *testing.T) {
	ctx := context.TODO()
	m := map[string]string{
		"version":           "2.0",
		"data_stream_id":    "test_stream",
		"data_stream_route": "test_route",
	}
	tuid := "1234"
	metadata.Cache.SetConfig("v2/test_stream-test_route.json", &validation.ManifestConfig{
		Copy: validation.CopyConfig{
			FilenameSuffix: metadata.FilenameSuffixUploadId,
		},
	})

	s, err := metadata.GetFilenameSuffix(ctx, m, tuid)
	if err != nil {
		t.Fatal(err)
	}
	if s != "_"+tuid {
		t.Error("expected upload ID suffix but get", s)
	}
}

func TestDeliveryFilenameSuffixNone(t *testing.T) {
	ctx := context.TODO()
	m := map[string]string{
		"version":           "2.0",
		"data_stream_id":    "test_stream",
		"data_stream_route": "test_route",
	}
	tuid := "1234"
	metadata.Cache.SetConfig("v2/test_stream-test_route.json", &validation.ManifestConfig{
		Copy: validation.CopyConfig{
			FilenameSuffix: "",
		},
	})

	s, err := metadata.GetFilenameSuffix(ctx, m, tuid)
	if err != nil {
		t.Fatal(err)
	}
	if s != "" {
		t.Error("expected empty suffix but get", s)
	}
}

func TestFileInfoNotFound(t *testing.T) {
	client := ts.Client()
	resp, err := client.Get(ts.URL + "/info/1234")

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Error("Expected 404 but got", resp.StatusCode)
	}
}

func TestMain(m *testing.M) {
	appConfig := appconfig.AppConfig{
		UploadConfigPath:      "../../upload-configs/",
		LocalFolderUploadsTus: "test/uploads",
		LocalReportsFolder:    "test/reports",
		LocalEventsFolder:     "test/events",
		LocalDEXFolder:        "test/dex",
		LocalEDAVFolder:       "test/edav",
		LocalRoutingFolder:    "test/routing",
		TusdHandlerBasePath:   "/files/",
	}

	testContext = context.Background()
	var testWaitGroup sync.WaitGroup
	defer testWaitGroup.Wait()
	//postProcessingChannel := make(chan event.FileReady)
	event.InitFileReadyChannel()
	testWaitGroup.Add(1)
	testListener := cli.MakeEventSubscriber(appConfig)
	go func() {
		cli.SubscribeToEvents(testContext, testListener)
		testWaitGroup.Done()
	}()

	serveHandler, err := cli.Serve(testContext, appConfig)
	if err != nil {
		log.Fatal(err)
	}

	ts = httptest.NewServer(serveHandler)
	testRes := m.Run()

	ts.Close()
	event.CloseFileReadyChannel()
	os.Exit(testRes)
}

type ReportSummary struct {
	Reports []reports.Report
	Count   int
}
type ReportFileSummary struct {
	Tuid      string
	Summaries map[string]ReportSummary
}

func readReportFile(tuid string) (ReportFileSummary, error) {
	summary := ReportFileSummary{
		Tuid:      tuid,
		Summaries: map[string]ReportSummary{},
	}

	f, err := os.Open("test/reports/" + tuid)
	if err != nil {
		return summary, fmt.Errorf("failed to open report file for %s; inner error %w", tuid, err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return summary, fmt.Errorf("failed to read report file %s; inner error %w", f.Name(), err)
	}

	trackedStages := []string{
		"dex-metadata-verify",
		"dex-metadata-transform",
		"dex-upload-status",
		"dex-upload-started",
		"dex-upload-complete",
		"dex-file-copy",
	}

	rScanner := bufio.NewScanner(strings.NewReader(string(b)))
	for rScanner.Scan() {
		rLine := rScanner.Text()
		rLineBytes := []byte(rLine)

		r, err := unmarshalReport(rLineBytes)
		if err != nil {
			return summary, err
		}

		for _, s := range trackedStages {
			if strings.Contains(rLine, s) {
				appendReport(summary, r)
			}
		}
	}

	return summary, nil
}

func unmarshalReport(bytes []byte) (reports.Report, error) {
	var r reports.Report
	err := json.Unmarshal(bytes, &r)
	return r, err
}

func appendReport(summary ReportFileSummary, r reports.Report) ReportFileSummary {
	stageName := r.StageInfo.Stage
	s, ok := summary.Summaries[stageName]
	if !ok {
		summary.Summaries[stageName] = ReportSummary{
			Count:   1,
			Reports: []reports.Report{r},
		}
		return summary
	}
	s.Count++
	s.Reports = append(s.Reports, r)
	return summary
}

func checkReportSummary(fileSummary ReportFileSummary, stageName string, expectedCount int) error {
	summary, ok := fileSummary.Summaries[stageName]

	if !ok {
		return fmt.Errorf("expected %d %s report but got none", expectedCount, stageName)
	} else if summary.Count != 1 {
		return fmt.Errorf("expected %d %s report but got %d", expectedCount, stageName, summary.Count)
	}

	return nil
}

func readEventFile(tuid string) ([]event.Event, error) {
	var events []event.Event
	f, err := os.Open("test/events/" + tuid)
	if err != nil {
		return nil, fmt.Errorf("failed to open event file file for %s; inner error %w", tuid, err)
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read event file %s; inner error %w", f.Name(), err)
	}

	rScanner := bufio.NewScanner(strings.NewReader(string(b)))
	for rScanner.Scan() {
		eLine := rScanner.Text()
		eLineBytes := []byte(eLine)

		var e event.Event
		err := json.Unmarshal(eLineBytes, &e)
		if err != nil {
			return nil, err
		}

		events = append(events, e)
	}

	return events, nil
}
