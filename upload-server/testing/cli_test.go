package testing

import (
	"encoding/json"
	"io"
	"log"
	"net/http/httptest"
	"os"
	"testing"

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
		if err != nil {
			t.Error(name, err)
		} else {

			if tuid != "" {

				f, err := os.Open("test/reports/" + tuid)
				if err != nil {
					t.Error(name, tuid, err)
				}

				r := &metadata.Report{}
				b, err := io.ReadAll(f)
				if err != nil {
					t.Fatal(name, tuid, err)
				}

				if err := json.Unmarshal(b, r); err != nil {
					t.Fatal(name, tuid, err)
				}
				if c.err != nil {
					if r.Content.(metadata.Content).Issues == nil {
						t.Error("expected reported issues but got none", name, tuid, r)
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
