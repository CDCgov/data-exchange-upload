package testing

import (
	"fmt"
	"os"

	"github.com/eventials/go-tus"
)

type testCase struct {
	metadata tus.Metadata
	err      error
}

var Cases = map[string]testCase{
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
			"sender_id":         "test",
			"data_producer_id":  "test",
			"jurisdiction":      "test",
			"received_filename": "test",
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
			"sender_id":              "test",
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
			"sender_id":         "test",
			"received_filename": "test",
			"message_type":      "ELR",
			"route":             "DAART",
			"jurisdiction":      "test",
		},
		nil,
	},
}

func RunTusTestCase(url string, testFile string, c testCase) error {
	f, err := os.Open(testFile)

	if err != nil {
		return fmt.Errorf("failed to open test file %w", err)
	}

	defer f.Close()

	// create the tus client.
	client, err := tus.NewClient(url, nil)
	if err != nil {
		return fmt.Errorf("failed to create test client %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat test file %w", err)
	}

	fingerprint := fmt.Sprintf("%s-%d-%s", fi.Name(), fi.Size(), fi.ModTime())
	c.metadata["filename"] = fi.Name()

	// create an upload from a file.
	upload := tus.NewUpload(f, fi.Size(), c.metadata, fingerprint)

	// create the uploader.
	uploader, err := client.CreateUpload(upload)
	if c.err != nil {
		if err == nil || c.err.Error() != err.Error() {
			return fmt.Errorf("error missmatch; got: %w wanted: %w", err, c.err)
		}
		return nil
	}

	if err != nil || uploader == nil {
		tErr, ok := err.(tus.ClientError)
		if ok {
			return fmt.Errorf("got a nil uploader or unexpected error %w, %s", err, string(tErr.Body))
		}
		return fmt.Errorf("got a nil uploader or unexpected error %w", err)
	}

	if err := uploader.Upload(); err != nil {
		return fmt.Errorf("failed to upload file %w", err)
	}
	return nil
}