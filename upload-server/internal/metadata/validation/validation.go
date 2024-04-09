package validation

import (
	"fmt"
)

type UploadConfig struct {
	Metadata MetadataConfig `json:"metadata_config"`
	Copy     CopyConfig     `json:"copy_config"`
}

type MetadataConfig struct {
	Version string        `json:"version"`
	Fields  []FieldConfig `json:"fields"`
}

type CopyConfig struct {
	FilenameSuffix  string   `json:"filename_suffix"`
	FolderStructure string   `json:"folder_structure"`
	Targets         []string `json:"targets"`
}

type FieldConfig struct {
	FieldName     string   `json:"field_name"`
	Required      bool     `json:"required"`
	Description   string   `json:"description"`
	AllowedValues []string `json:"allowed_values"`
}

type ErrorMissingRequired struct {
	Field string
}

func (e *ErrorMissingRequired) Error() string {
	return fmt.Sprintf("required field %s was missing", e.Field)
}

type ErrorMissing struct {
	field string
}

func (e *ErrorMissing) Error() string {
	return fmt.Sprintf("field %s was missing", e.field)
}

type ErrorNotAnAllowedValue struct {
	field string
	value string
}

func (e *ErrorNotAnAllowedValue) Error() string {
	return fmt.Sprintf("%s had disallowed value %s", e.field, e.value)
}

func (fc *FieldConfig) Validate(manifest map[string]string) error {
	value, ok := manifest[fc.FieldName]
	if !ok {
		if fc.Required {
			return &ErrorMissingRequired{Field: fc.FieldName}
		}
		return &ErrorMissing{field: fc.FieldName}
	}
	if len(fc.AllowedValues) > 0 {
		for _, allowed := range fc.AllowedValues {
			if allowed == value {
				return nil
			}
		}
		return &ErrorNotAnAllowedValue{field: fc.FieldName, value: value}
	}
	return nil
}

type ConfigLoader interface {
	LoadConfig(path string) ([]byte, error)
}

type ConfigGetter interface {
	GetConfig(ConfigLoader) (*MetadataConfig, error)
}
