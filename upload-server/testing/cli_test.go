package testing

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
)

var (
	ts *httptest.Server
)

func TestTus(t *testing.T) {
	url := ts.URL

	for name, c := range Cases {
		tuid, err := RunTusTestCase(url, "test/test.txt", c)
		time.Sleep(2 * time.Second) // TODO: Find a better way to wait for all the hooks to finish.

		if err != nil {
			t.Error(name, err)
		} else {

			if tuid != "" {
				f, err := os.Open("test/reports/" + tuid)
				if err != nil {
					t.Error(name, tuid, err)
				}

				metadataReportCount, uploadStatusReportCount, uploadStartedReportCount, uploadCompleteReportCount := 0, 0, 0, 0
				rMetadata, rUploadStatus := &metadata.Report{}, &metadata.Report{}
				b, err := io.ReadAll(f)
				if err != nil {
					t.Fatal(name, tuid, err)
				}

				rScanner := bufio.NewScanner(strings.NewReader(string(b)))
				for rScanner.Scan() {
					rLine := rScanner.Text()
					rLineBytes := []byte(rLine)
					if strings.Contains(rLine, "dex-metadata-verify") {
						// Processing a metadata verify report
						metadataReportCount++

						if err := json.Unmarshal(rLineBytes, rMetadata); err != nil {
							t.Fatal(name, tuid, err)
						}

						if rMetadata.DataStreamID == "" || rMetadata.DataStreamRoute == "" {
							t.Error("DataStreamID or DataStreamRoute is missing in metadata report", name, tuid)
						}

						continue
					}

					if strings.Contains(rLine, "dex-upload-status") {
						uploadStatusReportCount++

						if err := json.Unmarshal(rLineBytes, rUploadStatus); err != nil {
							t.Fatal(name, tuid, err)
						}

						if rUploadStatus.DataStreamID == "" || rUploadStatus.DataStreamRoute == "" {
							t.Error("DataStreamID or DataStreamRoute is missing in upload status report", name, tuid)
						}

						continue
					}

					if strings.Contains(rLine, "dex-upload-started") {
						uploadStartedReportCount++
						continue
					}

					if strings.Contains(rLine, "dex-upload-complete") {
						uploadCompleteReportCount++
						continue
					}
				}

				if metadataReportCount != 1 {
					t.Error("expected one metadata verify report but got", metadataReportCount)
				}

				if uploadStatusReportCount == 0 {
					t.Error("expected at least one upload status report count but got none")
				}

				if uploadStartedReportCount != 1 {
					t.Error("at least one upload started report count but got none", uploadStartedReportCount)
				}

				if uploadCompleteReportCount != 1 {
					t.Error("at least one upload complete report count but got none", uploadCompleteReportCount)
				}

				if c.err != nil {
					if rMetadata.Content.(metadata.MetaDataVerifyContent).Issues == nil {
						t.Error("expected reported issues but got none", name, tuid, rMetadata)
					}

					if rUploadStatus.Content.(metadata.UploadStatusContent).Offset != rUploadStatus.Content.(metadata.UploadStatusContent).Size {
						t.Error("expected latest status report to have equal offset and size but were different", name, tuid, rUploadStatus)
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
		TusdHandlerBasePath:   "/files/",
	}

	handler, err := cli.Serve(appConfig)
	if err != nil {
		log.Fatal(err)
	}
	ts = httptest.NewServer(handler)
	defer ts.Close()
	os.Exit(m.Run())
}
