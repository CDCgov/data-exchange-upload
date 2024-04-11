package validation

import "errors"

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

func (fc *FieldConfig) Validate(manifest map[string]string) error {
	value, ok := manifest[fc.FieldName]
	if !ok {
		if fc.Required {
			return errors.Join(ErrFailure, &ErrorMissing{Field: fc.FieldName})
		}
		return &ErrorMissing{Field: fc.FieldName}
	}
	if len(fc.AllowedValues) > 0 {
		for _, allowed := range fc.AllowedValues {
			if allowed == value {
				return nil
			}
		}
		return errors.Join(ErrFailure, &ErrorNotAnAllowedValue{field: fc.FieldName, value: value})
	}
	return nil
}

type ConfigLoader interface {
	LoadConfig(path string) ([]byte, error)
}

type ConfigLocation interface {
	Path() string
}
