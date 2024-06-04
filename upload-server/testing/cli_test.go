package testing

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

var (
	ts *httptest.Server
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
				f, err := os.Open("test/reports/" + tuid)
				if err != nil {
					t.Error(name, tuid, err)
				}

				metadataReportCount, uploadStatusReportCount, uploadStartedReportCount, uploadCompleteReportCount, metadataTransformReportCount := 0, 0, 0, 0, 0
				rMetadata, rUploadStatus := &models.Report{}, &models.Report{}
				b, err := io.ReadAll(f)
				if err != nil {
					t.Fatal(name, tuid, err)
				}

				rScanner := bufio.NewScanner(strings.NewReader(string(b)))
				for rScanner.Scan() {
					rLine := rScanner.Text()
					rLineBytes := []byte(rLine)

					if strings.Contains(rLine, "dex-metadata-transform") {
						metadataTransformReportCount++
						continue
					}

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

				if metadataTransformReportCount != 2 {
					t.Error("expected two metadata transform reports but got", metadataTransformReportCount)
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
					if rMetadata.Content.(models.MetaDataVerifyContent).Issues == nil {
						t.Error("expected reported issues but got none", name, tuid, rMetadata)
					}

					if rUploadStatus.Content.(models.UploadStatusContent).Offset != rUploadStatus.Content.(models.UploadStatusContent).Size {
						t.Error("expected latest status report to have equal offset and size but were different", name, tuid, rUploadStatus)
					}
				}

				// Post-processing
				// Check that the file exists in the dex checkpoint folder.
				if _, err := os.Stat("./test/dex/" + tuid); errors.Is(err, os.ErrNotExist) {
					t.Error("file was not copied to dex checkpoint for file", tuid)
				}
				// Also check that the .meta file exists in the dex folder.
				if _, err := os.Stat("./test/dex/" + tuid + ".meta"); errors.Is(err, os.ErrNotExist) {
					t.Error("meta file was not copied to dex checkpoint for file", tuid)
				}
				// Also check that the metadata in the .meta file is hydrated with v2 manifest fields.
				metaFile, err := os.Open("./test/dex/" + tuid + ".meta")
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
		LocalDEXFolder:        "test/dex",
		LocalEDAVFolder:       "test/edav",
		LocalROUTINGFolder:    "test/routing",
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
