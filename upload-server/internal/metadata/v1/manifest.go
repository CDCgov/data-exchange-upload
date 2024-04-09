package v1

import (
	"encoding/json"
	"errors"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/tus/tusd/v2/pkg/handler"

	"fmt"
)

type Config struct {
	MetaDestinationId string
	MetaExtEvent      string
}

func (c *Config) GetConfig(loader validation.ConfigLoader) (*validation.MetadataConfig, error) {
	path := fmt.Sprintf("%s/%s-%s.json", "v1", c.MetaDestinationId, c.MetaExtEvent)
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
	metaDestinationID, ok := manifest["meta_destination_id"]
	if !ok {
		return nil, errors.New("Missing meta_destination_id")
	}
	metaExtEvent, ok := manifest["meta_ext_event"]
	if !ok {
		return nil, errors.New("Missing meta_ext_event")
	}

	return &Config{
		MetaDestinationId: metaDestinationID,
		MetaExtEvent:      metaExtEvent,
	}, nil
}
