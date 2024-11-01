package testing

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/ui"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
)

var (
	ts            *httptest.Server
	testUIServer  *httptest.Server
	testContext   context.Context
	trackedStages = []string{
		reports.StageMetadataVerify,
		reports.StageMetadataTransform,
		reports.StageUploadCompleted,
		reports.StageUploadStarted,
		reports.StageUploadStatus,
		reports.StageFileCopy,
	}
)

const TestFolderUploadsTus = "uploads"
const TestEDAVFolder = "uploads/edav"
const TestEhdiFolder = "uploads/ehdi"
const TestEicrFolder = "uploads/eicr"
const TestEventsFolder = "uploads/events"
const TestNcirdFolder = "uploads/ncird"
const TestReportsFolder = "uploads/reports"

var AllTargets = map[string]string{
	"edav":  TestEDAVFolder,
	"ehdi":  TestEhdiFolder,
	"eicr":  TestEicrFolder,
	"ncird": TestNcirdFolder,
}

func TestTus(t *testing.T) {
	url := ts.URL

	for name, c := range Cases {
		tuid, err := RunTusTestCase(url, "test.txt", c)
		time.Sleep(2 * time.Second) // Hard delay to wait for all non-blocking hooks to finish.

		if err != nil {
			t.Error(name, err)
		} else {
			if tuid != "" {
				// Check that the .meta file exists in the dex folder.
				if _, err := os.Stat(TestFolderUploadsTus + "/" + tuid + ".meta"); errors.Is(err, os.ErrNotExist) {
					t.Error("meta file was not copied to dex checkpoint for file", tuid)
				}
				// .Check meta file

				// Check that the metadata in the .meta file is hydrated with v2 manifest fields.
				metaFile, err := os.Open(TestFolderUploadsTus + "/" + tuid + ".meta")
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

				appendedUid, ok := processedMeta["upload_id"]
				if !ok {
					t.Error("upload ID not appended to file metadata")
				} else if appendedUid != tuid {
					t.Error("appended upload ID did not match upload ID", appendedUid, tuid)
				}

				//translationFields := map[string]string{
				//	"meta_destination_id": "data_stream_id",
				//	"meta_ext_event":      "data_stream_route",
				//}
				//v, ok := c.metadata["version"]
				//
				//if !ok || v == "1.0" {
				//	for v1Key, v2Key := range translationFields {
				//		v1Val, ok := processedMeta[v1Key]
				//		if !ok {
				//			t.Error("malformed metadata; missing required field", v1Key, processedMeta)
				//		}
				//		v2Val, ok := processedMeta[v2Key]
				//		if !ok {
				//			t.Error("v1 metadata not hydrated; missing v2 field", v2Key)
				//		}
				//		if v1Val != v2Val {
				//			t.Error("v1 to v2 fields not properly translated", v1Val, v2Val)
				//		}
				//	}
				//}
				// .Check v2 hydration

				// Check that all of the report files were created
				expectedMetadataTransformReportCount := 2
				//if v, ok := c.metadata["version"]; !ok || v == "1.0" {
				//	expectedMetadataTransformReportCount = 3
				//}

				reportSummary, err := readReportFiles(tuid, trackedStages)
				if err != nil {
					t.Error("failed to read report file for", "tuid", tuid, err.Error())
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
				// .Check report files

				// Post-processing
				// Use the processedMeta data because the post-processing happens after hydration
				config, err := metadata.GetConfigFromManifest(testContext, processedMeta)
				if err != nil {
					t.Fatal(err)
				}

				events, err := readEventFile(tuid, event.FileReadyEventType)
				if err != nil {
					t.Error("no events found for tuid", "tuid", tuid)
				}
				if len(events) != len(config.Copy.Targets) {
					t.Errorf("expected %d file ready event(s) but got %d", len(config.Copy.Targets), len(events))
				}

				for target, path := range AllTargets {
					if slices.Contains(config.Copy.Targets, target) {
						// Check that the file exists in the target checkpoint folder
						if _, err := os.Stat(path + "/" + tuid); errors.Is(err, os.ErrNotExist) {
							t.Error("file was not copied to "+target+" checkpoint for file", tuid)
						}
					}
				}

			}

			t.Log("test case", name, "passed", tuid)
		}
	}
}

func TestRouteEndpoint(t *testing.T) {
	goodCase := "good"
	c, ok := Cases[goodCase]
	if !ok {
		t.Error("test case not found", "case", goodCase)
	}
	tuid, err := RunTusTestCase(ts.URL, "test.txt", c)
	time.Sleep(2 * time.Second) // Hard delay to wait for all non-blocking hooks to finish.

	if err != nil {
		t.Error("error", err)
	} else {
		if tuid == "" {
			t.Error("could not create tuid")
		} else {
			// Check that the file exists in the target checkpoint folder
			if _, err := os.Stat(TestEDAVFolder + "/" + tuid); errors.Is(err, os.ErrNotExist) {
				t.Error("file was not copied to edav checkpoint for file", tuid)
			}

			// Remove and re-route the file
			err = os.Remove(TestEDAVFolder + "/" + tuid)
			if err != nil {
				t.Error("failed to remove edav file for "+tuid, err.Error())
			}
			b := []byte(`{
				"target": "edav"
			}`)
			resp, err := http.Post(ts.URL+"/route/"+tuid, "application/json", bytes.NewBuffer(b))
			if err != nil {
				t.Error("failed to retry routing")
			}
			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				t.Error("expected 200 when retrying route but got", resp.StatusCode, string(b))
			}
			time.Sleep(100 * time.Millisecond) // Wait for new file ready event to be processed.
			if _, err := os.Stat(TestEDAVFolder + "/" + tuid); errors.Is(err, os.ErrNotExist) {
				t.Error("file was not copied to edav checkpoint when retry attempted for file", tuid)
			}
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

func TestRequiredUploadIdEndpoints(t *testing.T) {
	endpoints := []string{
		"/info",
		"/info/",
		"/route",
		"/route/",
	}
	client := ts.Client()
	for _, endpoint := range endpoints {
		resp, err := client.Get(ts.URL + endpoint)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 404 {
			t.Error("bad response for ", endpoint, resp.StatusCode)
		}
	}
}

func TestGetFileDeliveryPrefixDate(t *testing.T) {
	ctx := context.TODO()
	m := map[string]string{
		"data_stream_id":    "test-stream",
		"data_stream_route": "test-route",
	}
	metadata.Cache.SetConfig("test-stream_test-route.json", &validation.ManifestConfig{
		Copy: validation.CopyConfig{
			FolderStructure: metadata.FolderStructureDate,
		},
	})

	p, err := metadata.GetFilenamePrefix(ctx, m)
	if err != nil {
		t.Fatal(err)
	}
	prefixTokens := strings.Split(p, "/")
	if len(prefixTokens) != 4 {
		t.Error("prefix not properly formatted", p)
	}
	expectedFolderPrefix := m["data_stream_id"] + "-" + m["data_stream_route"]
	if prefixTokens[0] != expectedFolderPrefix {
		t.Error("prefix folder not properly formatted")
	}
}

func TestGetFileDeliveryPrefixRoot(t *testing.T) {
	ctx := context.TODO()
	m := map[string]string{
		"data_stream_id":    "test-stream",
		"data_stream_route": "test-route",
	}
	metadata.Cache.SetConfig("test-stream_test-route.json", &validation.ManifestConfig{
		Copy: validation.CopyConfig{
			FolderStructure: metadata.FolderStructureRoot,
		},
	})

	p, err := metadata.GetFilenamePrefix(ctx, m)
	if err != nil {
		t.Fatal(err)
	}
	expectedFolderPrefix := m["data_stream_id"] + "-" + m["data_stream_route"]

	if p != expectedFolderPrefix {
		t.Error("expected file delivery prefix to be folder prefix but was", p)
	}
}

func TestDeliveryFilenameSuffixUploadId(t *testing.T) {
	ctx := context.TODO()
	m := map[string]string{
		"data_stream_id":    "test-stream",
		"data_stream_route": "test-route",
	}
	tuid := "1234"
	metadata.Cache.SetConfig("test-stream_test-route.json", &validation.ManifestConfig{
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
		"data_stream_id":    "test-stream",
		"data_stream_route": "test-route",
	}
	tuid := "1234"
	metadata.Cache.SetConfig("test-stream_test-route.json", &validation.ManifestConfig{
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

func TestRouteBadRequest(t *testing.T) {
	client := ts.Client()
	resp, err := client.Get(ts.URL + "/route/1234")

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Error("Expected 400 but got", resp.StatusCode)
	}
}

func TestRouteInvalidBody(t *testing.T) {
	client := ts.Client()
	b := []byte("blah")
	resp, err := client.Post(ts.URL+"/route/1234", "application/json", bytes.NewBuffer(b))

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Error("Expected 400 but got", resp.StatusCode)
	}
}

func TestRouteInvalidTarget(t *testing.T) {
	client := ts.Client()
	b := []byte(`{
		"target": "blah"
	}`)
	path := ts.URL + "/route/1234"
	resp, err := client.Post(path, "application/json", bytes.NewBuffer(b))

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Error("Expected 404 but got", resp.StatusCode, path)
	}
}

func TestRouteFileNotFound(t *testing.T) {
	client := ts.Client()
	b := []byte(`{
		"target": "edav"
	}`)
	resp, err := client.Post(ts.URL+"/route/1234", "application/json", bytes.NewBuffer(b))

	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Error("Expected 404 but got", resp.StatusCode)
	}
}

// UI Tests
func TestLandingPage(t *testing.T) {
	client := testUIServer.Client()
	resp, err := client.Get(testUIServer.URL)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Expected 200 but got", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Error("Expected html content but got", ct)
	}
}

func TestManifestPageManifestNoQueryParams(t *testing.T) {
	client := testUIServer.Client()
	resp, err := client.Get(testUIServer.URL + "/manifest")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Error("Expected 404 but got", resp.StatusCode)
	}
}

func TestManifestPageManifestNotFound(t *testing.T) {
	client := testUIServer.Client()
	resp, err := client.Get(testUIServer.URL + "/manifest?data_stream_id=invalid&data_stream_route=invalid")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Error("Expected 404 but got", resp.StatusCode)
	}
}

func TestManifestPageValidDestination(t *testing.T) {
	client := testUIServer.Client()
	resp, err := client.Get(testUIServer.URL + "/manifest?data_stream_id=dextesting&data_stream_route=testevent1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Expected 200 but got", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Error("Expected html content but got", ct)
	}
}

func TestUploadPageEmptyBodyNotFound(t *testing.T) {
	client := testUIServer.Client()
	manifestForm := url.Values{}
	body := strings.NewReader(manifestForm.Encode())
	resp, err := client.Post(testUIServer.URL+"/upload", "application/x-www-form-urlencoded", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Error("Expected 400 but got", resp.StatusCode)
	}
}

func TestUploadPageInvalidManifestBadRequest(t *testing.T) {
	client := testUIServer.Client()
	manifestForm := url.Values{
		"data_stream_id":    {"dextesting"},
		"data_stream_route": {"testevent1"},
	}
	body := strings.NewReader(manifestForm.Encode())
	resp, err := client.Post(testUIServer.URL+"/upload", "application/x-www-form-urlencoded", body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		t.Error("Expected 400 but got", resp.StatusCode)
	}
}

func TestUploadPageRedirectStatusPage(t *testing.T) {
	didRedirect := false
	var redirectUrl *url.URL
	client := testUIServer.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		didRedirect = true
		redirectUrl = req.URL
		return nil
	}
	manifestForm := url.Values{
		"data_stream_id":    {"dextesting"},
		"data_stream_route": {"testevent1"},
		"sender_id":         {"test sender ID"},
		"data_producer_id":  {"test data producer ID"},
		"jurisdiction":      {"test jur"},
		"received_filename": {"test.txt"},
	}
	body := strings.NewReader(manifestForm.Encode())
	resp, err := client.Post(testUIServer.URL+"/upload", "application/x-www-form-urlencoded", body)
	if err != nil {
		t.Fatal(err)
	}
	if !didRedirect {
		t.Error("Expected to redirect but did not")
	}
	if resp.StatusCode != 200 {
		t.Error("Expected 200 but got", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct != "text/html; charset=utf-8" {
		t.Error("Expected html content but got", ct)
	}
	if resp.Request.URL.String() != redirectUrl.String() {
		t.Errorf("Expected to be redirected to %s but was redirected to %s", redirectUrl.String(), resp.Request.URL.String())
	}
}

func TestStatusPageUploadNotFoundRedirect(t *testing.T) {
	didRedirect := false
	client := testUIServer.Client()
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		didRedirect = true
		return nil
	}
	resp, err := client.Get(testUIServer.URL + "/status/1234")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Error("Expected 200 but got", resp.StatusCode)
	}
	if !didRedirect {
		t.Error("Expected to redirect but did not")
	}
}

func TestMain(m *testing.M) {
	oauthConfig := appconfig.OauthConfig{
		AuthEnabled:      false,
		IssuerUrl:        "",
		RequiredScopes:   "",
		IntrospectionUrl: "",
	}
	appConfig := appconfig.AppConfig{
		UploadConfigPath:      "../../upload-configs/",
		LocalFolderUploadsTus: "./" + TestFolderUploadsTus,
		LocalReportsFolder:    "./" + TestReportsFolder,
		LocalEventsFolder:     "./" + TestEventsFolder,
		DeliveryConfigFile:    "./delivery.yml",
		TusdHandlerBasePath:   "/files/",
		OauthConfig:           &oauthConfig,
	}
	appconfig.LoadedConfig = &appConfig

	testContext = context.Background()
	var testWaitGroup sync.WaitGroup
	defer testWaitGroup.Wait()

	err := delivery.RegisterAllSourcesAndDestinations(testContext, appConfig)
	event.InitFileReadyChannel()
	testWaitGroup.Add(1)
	err = cli.InitReporters(testContext, appConfig)
	defer reports.CloseAll()
	err = event.InitFileReadyPublisher(testContext, appConfig)
	defer event.FileReadyPublisher.Close()
	testListener, err := cli.NewEventSubscriber[*event.FileReady](testContext, appConfig)
	go func() {
		testListener.Listen(testContext, postprocessing.ProcessFileReadyEvent)
		testWaitGroup.Done()
	}()

	serveHandler, err := cli.Serve(testContext, appConfig)
	if err != nil {
		log.Fatal(err)
	}

	ts = httptest.NewServer(serveHandler)

	// Start ui server
	uiHandler := ui.GetRouter(ts.URL+appConfig.TusdHandlerBasePath, ts.URL+appConfig.TusdHandlerInfoPath, ts.URL+appConfig.TusdHandlerBasePath)
	testUIServer = httptest.NewServer(uiHandler)

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

func readReportFiles(tuid string, stages []string) (ReportFileSummary, error) {
	summary := ReportFileSummary{
		Tuid:      tuid,
		Summaries: map[string]ReportSummary{},
	}

	for _, stage := range stages {
		filename := tuid + event.TypeSeparator + stage
		f, err := os.Open(TestReportsFolder + "/" + filename)
		if err != nil {
			return summary, fmt.Errorf("failed to open report file for %s; inner error %w", filename, err)
		}
		b, err := io.ReadAll(f)
		if err != nil {
			return summary, fmt.Errorf("failed to read report file %s; inner error %w", f.Name(), err)
		}

		rScanner := bufio.NewScanner(strings.NewReader(string(b)))
		for rScanner.Scan() {
			rLine := rScanner.Text()
			rLineBytes := []byte(rLine)

			r, err := unmarshalReport(rLineBytes)
			if err != nil {
				return summary, err
			}

			appendReport(summary, r)
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
	stageName := r.StageInfo.Action
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

func readEventFile(tuid string, eType string) ([]event.Event, error) {
	var events []event.Event
	f, err := os.Open(TestEventsFolder + "/" + tuid + event.TypeSeparator + eType)
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
