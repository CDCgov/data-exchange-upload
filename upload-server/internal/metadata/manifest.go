package metadata

import (
	"errors"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/tus/tusd/v2/pkg/handler"

	"fmt"
)

type ConfigIdentification struct {
	DataStreamID    string
	DataStreamRoute string
}

func (c *ConfigIdentification) Path() string {
	path := fmt.Sprintf("%s_%s.json", c.DataStreamID, c.DataStreamRoute)
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

	return &ConfigIdentification{
		DataStreamID:    dataStreamID,
		DataStreamRoute: dataStreamRoute,
	}, nil
}
