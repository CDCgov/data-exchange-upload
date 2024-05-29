package v1

import (
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/tus/tusd/v2/pkg/handler"

	"errors"
	"fmt"
)

type Config struct {
	MetaDestinationId string
	MetaExtEvent      string
}

func (c *Config) Path() string {
	path := fmt.Sprintf("%s/%s-%s.json", "v1", c.MetaDestinationId, c.MetaExtEvent)
	return path

}

func NewFromManifest(manifest handler.MetaData) (validation.ConfigLocation, error) {
	metaDestinationID, ok := manifest["meta_destination_id"]
	if !ok {
		return nil, errors.Join(validation.ErrFailure, &validation.ErrorMissing{Field: "meta_destination_id"})
	}
	metaExtEvent, ok := manifest["meta_ext_event"]
	if !ok {
		return nil, errors.Join(validation.ErrFailure, &validation.ErrorMissing{Field: "meta_ext_event"})
	}

	return &Config{
		MetaDestinationId: metaDestinationID,
		MetaExtEvent:      metaExtEvent,
	}, nil
}

func Hydrate(m map[string]string, config *validation.ManifestConfig) map[string]string {
	for _, field := range config.Metadata.Fields {
		if v, ok := m[field.FieldName]; ok {
			m[field.CompatFieldName] = v
		}
	}
	return m
}
