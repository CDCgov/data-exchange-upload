package testing

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
	"github.com/eventials/go-tus"
)

type testCase struct {
	metadata   tus.Metadata
	err        error
	deliveries []info.FileDeliveryStatus
}

var Cases = map[string]testCase{
	"good": {
		tus.Metadata{
			"meta_destination_id": "dextesting",
			"meta_ext_event":      "testevent1",
		},
		nil,
		[]info.FileDeliveryStatus{
			{},
			{},
			{},
			{},
		},
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
		[]info.FileDeliveryStatus{
			{},
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
		[]info.FileDeliveryStatus{
			{},
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
		[]info.FileDeliveryStatus{
			{},
		},
	},
	"v2 good": {
		tus.Metadata{
			"version":           "2.0",
			"data_stream_id":    "dextesting",
			"data_stream_route": "testevent1",
			"sender_id":         "test",
			"data_producer_id":  "test",
			"jurisdiction":      "test",
			"received_filename": "test",
		},
		nil,
		[]info.FileDeliveryStatus{
			{},
			{},
			{},
			{},
		},
	},
	"eicr v2 bad (missing things)": {
		tus.Metadata{
			"version":           "2.0",
			"data_stream_id":    "eicr",
			"data_stream_route": "fhir",
			"sender_id":         "APHL",
			"original_filename": "test",
			"message_type":      "bad",
		},
		tus.ClientError{
			Code: 400,
		},
		[]info.FileDeliveryStatus{
			{},
		},
	},
	"eicr v2 good": {
		tus.Metadata{
			"version":                 "2.0",
			"data_stream_id":          "eicr",
			"data_stream_route":       "fhir",
			"data_producer_id":        "test",
			"sender_id":               "APHL",
			"received_filename":       "test",
			"jurisdiction":            "test",
			"meta_ext_objectkey":      "test",
			"meta_ext_file_timestamp": "test",
		},
		nil,
		[]info.FileDeliveryStatus{
			{},
		},
	},
	"ndlp v1 good": {
		tus.Metadata{
			"meta_destination_id":       "ndlp",
			"meta_ext_event":            "covidallmonthlyvaccination",
			"meta_ext_filename":         "test",
			"meta_ext_objectkey":        "test",
			"meta_ext_source":           "IZGW",
			"meta_ext_sourceversion":    "V2023-09-01",
			"meta_ext_submissionperiod": "test",
			"meta_username":             "test",
			"meta_ext_entity":           "AKA",
		},
		nil,
		[]info.FileDeliveryStatus{
			{},
		},
	},
}

func RunTusTestCase(url string, testFile string, c testCase) (string, error) {
	var tuid string
	f, err := os.Open(testFile)

	if err != nil {
		return "", fmt.Errorf("failed to open test file %w", err)
	}

	defer f.Close()
	paths := []string{"/files/", "/files"}

	for _, path := range paths {
		// create the tus client.
		client, err := tus.NewClient(url+path, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create test client %w", err)
		}

		fi, err := f.Stat()
		if err != nil {
			return "", fmt.Errorf("failed to stat test file %w", err)
		}

		fingerprint := fmt.Sprintf("%s-%d-%s", fi.Name(), fi.Size(), fi.ModTime())
		c.metadata["filename"] = fi.Name()

		// create an upload from a file.
		upload := tus.NewUpload(f, fi.Size(), c.metadata, fingerprint)

		// create the uploader.
		uploader, err := client.CreateUpload(upload)
		if c.err != nil {
			if err == nil || c.err.Error() != err.Error() {
				return "", fmt.Errorf("error missmatch; got: %w wanted: %w", err, c.err)
			}
			return "", nil
		}

		if err != nil || uploader == nil {
			tErr, ok := err.(tus.ClientError)
			if ok {
				return "", fmt.Errorf("got a nil uploader or unexpected error %w, %s", err, string(tErr.Body))
			}
			return "", fmt.Errorf("got a nil uploader or unexpected error %w", err)
		}

		if err := uploader.Upload(); err != nil {
			return "", fmt.Errorf("failed to upload file %w", err)
		}

		tuid = filepath.Base(uploader.Url())

		time.Sleep(1 * time.Second)
		// check the file
		resp, err := http.Get(url + "/info/" + tuid)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("failed to get upload info %s", resp.Status)
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		infoJson := &info.InfoResponse{}
		if err := json.Unmarshal(body, infoJson); err != nil {
			return "", err
		}

		_, ok := infoJson.FileInfo["size_bytes"]
		if !ok {
			return "", fmt.Errorf("invalid info response: %s", infoJson)
		}

		// check hydrated manifest fields
		_, ok = infoJson.Manifest["dex_ingest_datetime"]
		if !ok {
			return "", fmt.Errorf("invalid file manifest: %s", infoJson.Manifest)
		}

		if len(infoJson.Deliveries) != len(c.deliveries) {
			return "", fmt.Errorf("incorrect deliveries: %+v %+v", infoJson.Deliveries, c.deliveries)
		}
	}

	return tuid, nil
}
