package appconfig

import (
	"context"

	"github.com/sethvargo/go-envconfig"
) // .import

type AppConfig struct {

	// App and for Logger
	System        string `env:"SYSTEM, required"`
	DexProduct    string `env:"DEX_PRODUCT, required"`
	DexApp        string `env:"DEX_APP, required"`
	LoggerDebugOn bool   `env:"LOGGER_DEBUG_ON"`

	// Server
	ServerPort string `env:"SERVER_PORT, required"`

	// Metadata
	MetadataVersions string `env:"METADATA_VERSIONS, required"`

	// Metadata v1
	AllowedDestAndEventsPath string `env:"ALLOWED_DEST_AND_EVENTS_PATH, required"`
	DefinitionsPath          string `env:"DEFINITIONS_PATH, required"`
	UploadConfigPath         string `env:"UPLOAD_CONFIG_PATH, required"`

	// Local
	LocalFolderUploads	string `env:"LOCAL_FOLDER_UPLOADS, required"`

	// Azure
	AzStorageName         string `env:"AZ_STORAGE_NAME"`
	AzStorageKey          string `env:"AZ_STORAGE_KEY"`
	AzContainerName       string `env:"AZ_CONTAINER_NAME"`
	AzContainerEndpoint   string `env:"AZ_CONTAINER_ENDPOINT"`
	AzContainerAccessType string `env:"AZ_CONTAINER_ACCESS_TYPE"`
} // .AppConfig

// ParseConfig loads app configuration based on environment variables and returns AppConfig struct
func ParseConfig(ctx context.Context) (AppConfig, error) {

	var ac AppConfig
	if err := envconfig.Process(ctx, &ac); err != nil {
		return AppConfig{}, err
	} // .if

	return ac, nil
} // .ParseConfig
