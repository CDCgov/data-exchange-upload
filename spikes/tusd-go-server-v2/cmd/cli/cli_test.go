package cli

import (
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	tus "github.com/eventials/go-tus"
	"github.com/tus/tusd/v2/pkg/filestore"
	"github.com/tus/tusd/v2/pkg/memorylocker"
)

var (
	appConfig = appconfig.AppConfig{
		AllowedDestAndEventsPath: "../../configs/allowed_destination_and_events.json",
		DefinitionsPath:          "../../configs/file-hooks/metadata-verify/",
		UploadConfigPath:         "../../configs/upload-configs/",
		HydrateV1ConfigPath:      "../../configs/upload-configs/v2/",
	}
	store = filestore.FileStore{
		Path: "test/uploads",
	}
	locker = memorylocker.New()
)

func init() {
	_, err := metadatav1.LoadOnce(appConfig)
	if err != nil {
		log.Fatal(err)
	}
}

type testCase struct {
	metadata tus.Metadata
	err      error
}

func TestTus(t *testing.T) {
	hookHandler := PrebuiltHooks()

	handlerTusd, err := handlertusd.New(store, locker, hookHandler, "/")

	if err != nil {
		t.Error(err)
	} // .handlerTusd

	ts := httptest.NewServer(handlerTusd)
	defer ts.Close()

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
		f, err := os.Open("test/test.txt")

		if err != nil {
			t.Fatal(err)
		}

		defer f.Close()

		// create the tus client.
		client, err := tus.NewClient(ts.URL, nil)
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
		if c.err != nil {
			if c.err.Error() != err.Error() {
				t.Error("error missmatch", "got", err, "wanted", c.err)
			}
		}

		if uploader != nil {
			// start the uploading process.
			if err := uploader.Upload(); err != nil {
				t.Error(err)
			}
		}
	}

	//TODO assert that expected results are in the right place
}
