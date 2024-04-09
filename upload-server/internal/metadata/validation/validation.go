package validation

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

type ConfigLoader interface {
	LoadConfig(path string) ([]byte, error)
}

type ConfigGetter interface {
	GetConfig(ConfigLoader) (*MetadataConfig, error)
}
