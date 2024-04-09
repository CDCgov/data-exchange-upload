package cli

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	tus "github.com/eventials/go-tus"
)

var (
	ts *httptest.Server
)

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
		"missing meta_destination_id": {
			tus.Metadata{
				"bad_key":        "dextesting",
				"meta_ext_event": "testevent1",
			},
			tus.ClientError{
				Code: 400,
				Body: []byte("meta_destination_id not found in manifest"),
			},
		},
		"missing meta_ext_event": {
			tus.Metadata{
				"meta_destination_id": "dextesting",
				"bad_key":             "testevent1",
			},
			tus.ClientError{
				Code: 400,
				Body: []byte("meta_ext_event not found in manifest"),
			},
		},
		"unkown meta_ext_event": {
			tus.Metadata{
				"meta_destination_id": "dextesting",
				"meta_ext_event":      "nonsense",
			},
			tus.ClientError{
				Code: 400,
				Body: []byte("configuration not found"),
			},
		},
		"v2 good": {
			tus.Metadata{
				"version":           "2.0",
				"data_stream_id":    "dextesting",
				"data_stream_route": "testevent1",
			},
			nil,
		},
		"daart good": {
			tus.Metadata{
				"meta_destination_id":    "daart",
				"meta_ext_event":         "hl7",
				"original_filename":      "test",
				"message_type":           "ELR",
				"route":                  "DAART",
				"reporting_jurisdiction": "test",
			},
			nil,
		},
		"daart bad": {
			tus.Metadata{
				"meta_destination_id":    "daart",
				"meta_ext_event":         "hl7",
				"original_filename":      "test",
				"message_type":           "bad",
				"reporting_jurisdiction": "test",
				"route":                  "DAART",
			},
			tus.ClientError{
				Code: 400,
			},
		},
		"daart v2 bad (missing things)": {
			tus.Metadata{
				"version":                "2.0",
				"data_stream_id":         "daart",
				"data_stream_route":      "hl7",
				"original_filename":      "test",
				"message_type":           "bad",
				"route":                  "DAART",
				"reporting_jurisdiction": "test",
			},
			tus.ClientError{
				Code: 400,
			},
		},
		"daart v2 good": {
			tus.Metadata{
				"version":           "2.0",
				"data_stream_id":    "daart",
				"data_stream_route": "hl7",
				"data_producer_id":  "test",
				"received_filename": "test",
				"message_type":      "ELR",
				"route":             "DAART",
				"jurisdiction":      "test",
			},
			nil,
		},
	}
	for name, c := range cases {
		t.Log(name)
		testTus(c, t)
	}
}

func testTus(c testCase, t *testing.T) {
	f, err := os.Open("test/test.txt")

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
	if c.err != nil && (err == nil || c.err.Error() != err.Error()) {
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
	appConfig := appconfig.AppConfig{
		AllowedDestAndEventsPath: "../../configs/allowed_destination_and_events.json",
		DefinitionsPath:          "../../configs/file-hooks/metadata-verify/",
		UploadConfigPath:         "../../configs/upload-configs/",
		HydrateV1ConfigPath:      "../../configs/upload-configs/v2/",
		LocalFolderUploadsTus:    "test/uploads",
		TusdHandlerBasePath:      "/files/",
	}

	handler, err := Serve(appConfig)
	if err != nil {
		log.Fatal(err)
	}
	ts = httptest.NewServer(handler)
	defer ts.Close()
	os.Exit(m.Run())
}
