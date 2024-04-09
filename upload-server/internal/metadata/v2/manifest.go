package v2

import (
	"encoding/json"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/tus/tusd/v2/pkg/handler"

	"fmt"
)

type Config struct {
	DataStreamID    string
	DataStreamRoute string
}

func (c *Config) GetConfig(loader validation.ConfigLoader) (*validation.MetadataConfig, error) {
	path := fmt.Sprintf("%s/%s-%s.json", "v2", c.DataStreamID, c.DataStreamRoute)
	// load the file
	b, err := loader.LoadConfig(path)
	if err != nil {
		return nil, err
	}
	config := &validation.UploadConfig{}
	if err := json.Unmarshal(b, config); err != nil {
		return nil, err
	}
	return &config.Metadata, nil

}

func NewFromManifest(manifest handler.MetaData) (*Config, error) {
	dataStreamID, ok := manifest["data_stream_id"]
	if !ok {
		return nil, &validation.ErrorMissingRequired{Field: "data_stream_id"}
	}
	dataStreamRoute, ok := manifest["data_stream_route"]
	if !ok {
		return nil, &validation.ErrorMissingRequired{Field: "data_stream_route"}
	}

	return &Config{
		DataStreamID:    dataStreamID,
		DataStreamRoute: dataStreamRoute,
	}, nil
}
