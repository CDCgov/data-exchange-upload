package integration_test

import (
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/eventials/go-tus"
	"github.com/joho/godotenv"
)

const IntegrationConfigPath = "./integration.env"

var ts *httptest.Server

type testCase struct {
	metadata tus.Metadata
	err      error
}

func TestTus(t *testing.T) {

	cases := map[string]testCase{
		"good": {
			tus.Metadata{
				"meta_destination_id": "dextesting",
				"meta_ext_event":      "testevent1",
			},
			nil,
		},
		"bad": {
			tus.Metadata{
				"bad_key":        "dextesting",
				"meta_ext_event": "testevent1",
			},
			tus.ClientError{
				Code: 400,
				Body: []byte("meta_destination_id not found in manifest"),
			},
		},
	}
	for _, c := range cases {
		testTus(c, t)
	}
}

func testTus(c testCase, t *testing.T) {
	f, err := os.Open("../cmd/cli/test/test.txt")

	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	// create the tus client.
	client, err := tus.NewClient(ts.URL+"/files/", nil)
	if err != nil {
		t.Error(err)
	}

	fi, err := f.Stat()
	if err != nil {
		t.Error(err)
	}

	fingerprint := fmt.Sprintf("%s-%d-%s", fi.Name(), fi.Size(), fi.ModTime())
	c.metadata["filename"] = fi.Name()

	// create an upload from a file.
	upload := tus.NewUpload(f, fi.Size(), c.metadata, fingerprint)

	// create the uploader.
	uploader, err := client.CreateUpload(upload)
	if c.err != nil && c.err.Error() != err.Error() {
		t.Error("error missmatch", "got", err, "wanted", c.err)
	}

	if uploader == nil {
		return
	}

	if err := uploader.Upload(); err != nil {
		t.Error(err)
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
	if _, ok := os.LookupEnv("UPLOAD_INTEGRATION_TEST"); !ok {
		log.Println("Not running integration tests")
		return
	}
	ctx := context.Background()
	if err := godotenv.Load(IntegrationConfigPath); err != nil {
		log.Fatal(err)
	} // .if
	// ------------------------------------------------------------------
	// parse and load config from os exported
	// ------------------------------------------------------------------
	appConfig, err := appconfig.ParseConfig(ctx)
	if err != nil {
		log.Fatal(err)
	} // .if

	handler, err := cli.Serve(appConfig)
	if err != nil {
		log.Fatal(err)
	}
	ts = httptest.NewServer(handler)
	defer ts.Close()
	os.Exit(m.Run())
}
