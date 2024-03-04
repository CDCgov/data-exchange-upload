package metadatav1

// struct definitions for metadata v1

type MetadataV1 struct {
	AllowedDestAndEvents []AllowedDestAndEvents

	Definitions AllDefinitions

	UploadConfigs AllUploadConfigs
} // .MetadataV1

// -----------------------------------------------------
// Allowed destination and events
// -----------------------------------------------------
type AllowedDestAndEvents struct {
	DestinationId string      `json:"destination_id"`
	ExtEvents     []ExtEvents `json:"ext_events"`
} // .AllowedDestAndEvents

type ExtEvents struct {
	Name               string       `json:"name"`
	DefinitionFileName string       `json:"definition_filename"`
	CopyTargets        []CopyTarget `json:"copy_targets"`
} // .extEvents

type CopyTarget struct {
	Target string `json:"target"`
} // .copyTarget

// -----------------------------------------------------
// Definitions
// -----------------------------------------------------
type AllDefinitions map[string][]Definition

type Definition struct {
	SchemaVersion string  `json:"schema_version"`
	Fields        []Field `json:"fields"`
} // .Definition

type Field struct {
	FieldName     string   `json:"fieldname"`
	AllowedValues []string `json:"allowed_values"`
	Required      string   `json:"required"`
	Description   string   `json:"description"`
} // .Field

// -----------------------------------------------------
// TUpload config
// -----------------------------------------------------
type AllUploadConfigs map[string]UploadConfig

type UploadConfig struct {
	FileNameMetadataField string `json:"FilenameMetadataField"`
	FileNameSuffix        string `json:"FilenameSuffix"`
	FolderStructure       string `json:"FolderStructure"`
} // .UploadConfig
