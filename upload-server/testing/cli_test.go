package testing

import (
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

var (
	ts *httptest.Server
)

func TestTus(t *testing.T) {
	url := ts.URL + "/files/"
	for name, c := range Cases {
		if err := RunTusTestCase(url, "test/test.txt", c); err != nil {
			t.Error(name, err)
		} else {
			t.Log("test case", name, "passed")
		}
	}
}

func TestWellKnownEndpoints(t *testing.T) {
	endpoints := []string{
		"/",
		"/health",
		"/version",
		"/metadata",
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

func TestMain(m *testing.M) {
	appConfig := appconfig.AppConfig{
		UploadConfigPath:      "../../upload-configs/",
		LocalFolderUploadsTus: "test/uploads",
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