package v2

import (
	"errors"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/tus/tusd/v2/pkg/handler"

	"fmt"
)

type Config struct {
	DataStreamID    string
	DataStreamRoute string
}

func (c *Config) Path() string {
	path := fmt.Sprintf("%s/%s-%s.json", "v2", c.DataStreamID, c.DataStreamRoute)
	return path
}

func NewFromManifest(manifest handler.MetaData) (validation.ConfigLocation, error) {
	dataStreamID, ok := manifest["data_stream_id"]
	if !ok {
		return nil, errors.Join(validation.ErrFailure, &validation.ErrorMissing{Field: "data_stream_id"})
	}
	dataStreamRoute, ok := manifest["data_stream_route"]
	if !ok {
		return nil, errors.Join(validation.ErrFailure, &validation.ErrorMissing{Field: "data_stream_route"})
	}

	return &Config{
		DataStreamID:    dataStreamID,
		DataStreamRoute: dataStreamRoute,
	}, nil
}
