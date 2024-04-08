package metadatav1

// MetadataV1 struct definitions for metadata v1
type MetadataV1 struct {
	AllowedDestAndEvents []AllowedDestAndEvents `json:"allowed_destination_and_events"`

	Definitions AllDefinitions `json:"all_definitions"`

	UploadConfigs AllUploadConfigs `json:"all_upload_configs"`

	// maps to do fast checks in hooks
	DestIdsEventsNameMap   DestIdsEventsNameMap   `json:"-"` // no need to expose on http endpoint but could for visibility if needed
	DestIdEventFileNameMap DestIdEventFileNameMap `json:"-"` // same above

	HydrateV1ConfigsMap HydrateV1ConfigsMap `json:"hydrate_v1_config_map"`
} // .MetadataV1

type DestIdsEventsNameMap map[string][]string
type DestIdEventFileNameMap map[string]string

type HydrateV1ConfigsMap map[string]HydrateV1Config

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
// Upload config
// -----------------------------------------------------
type AllUploadConfigs map[string]UploadConfig

type UploadConfig struct {
	FileNameMetadataField string `json:"FilenameMetadataField"`
	FileNameSuffix        string `json:"FilenameSuffix"`
	FolderStructure       string `json:"FolderStructure"`
} // .UploadConfig

// -----------------------------------------------------
// V1 to V2 Hydrate
// -----------------------------------------------------
type HydrateV1Config struct {
	FilenameSuffix  string `json:"filename_suffix"`
	FolderStructure string `json:"folder_structure"`
	MetadataConfig  struct {
		Version string `json:"version"`
		Fields  []struct {
			FieldName       string `json:"field_name"`
			CompatFieldName string `json:"compat_field_name"`
			DefaultValue    string `json:"default_value"`
		} `json:"fields"`
	} `json:"metadata_config"`
} // .HydrateV1Config
