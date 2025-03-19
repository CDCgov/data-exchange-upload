package delivery_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
)

func TestGetDeliveredFilename(t *testing.T) {
	type testCase struct {
		ctx          context.Context
		tuid         string
		pathTemplate string
		manifest     map[string]string
		err          error
		result       string
	}
	ctx := context.Background()
	testTime := time.Date(2020, time.April, 11, 15, 12, 30, 50, time.UTC)

	testCases := []testCase{
		{
			ctx,
			"bogus-id",
			"routine-immunization-zip/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
			map[string]string{
				"version":             "2.0",
				"data_stream_id":      "dextesting",
				"data_stream_route":   "testevent1",
				"dex_ingest_datetime": testTime.Format(time.RFC3339Nano),
				"filename":            "test.txt",
			},
			nil,
			"routine-immunization-zip/2020/04/11/test.txt",
		},
		{
			ctx,
			"bogus-id",
			"routine-immunization-zip/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
			map[string]string{
				"version":             "2.0",
				"data_stream_id":      "dextesting",
				"data_stream_route":   "testevent1",
				"dex_ingest_datetime": "bogus time",
				"filename":            "test.txt",
			},
			delivery.ErrBadIngestTimestamp,
			"",
		},
		{
			ctx,
			"bogus-id",
			"routine-immunization-zip/{{.Year}}/{{.Month}}/{{.Day}}/{{.Filename}}",
			map[string]string{
				"version":           "2.0",
				"data_stream_id":    "dextesting",
				"data_stream_route": "testevent1",
				"filename":          "test.txt",
			},
			delivery.ErrBadIngestTimestamp,
			"",
		},
	}

	for i, c := range testCases {
		res, err := delivery.GetDeliveredFilename(c.ctx, c.tuid, c.pathTemplate, c.manifest)
		if res != c.result {
			t.Errorf("missmatched results for test case %d: got %s expected %s", i, res, c.result)
		}
		if !errors.Is(err, c.err) || (c.err == nil && err != nil) {
			t.Errorf("missmatched errors for test case %d: got %+v expected %+v", i, err, c.err)
		}
	}
}
