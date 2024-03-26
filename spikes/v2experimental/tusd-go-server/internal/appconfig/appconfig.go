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
	ServerPort  string `env:"SERVER_PORT, required"`
	Environment string `env:"ENVIRONMENT, required"`

	// Metadata
	MetadataVersions string `env:"METADATA_VERSIONS, required"`

	// Metadata v1
	AllowedDestAndEventsPath string `env:"ALLOWED_DEST_AND_EVENTS_PATH, required"`
	DefinitionsPath          string `env:"DEFINITIONS_PATH, required"`
	UploadConfigPath         string `env:"UPLOAD_CONFIG_PATH, required"`

	// Local folder path e.g. ../uploads
	LocalFolderUploadsTus string `env:"LOCAL_FOLDER_UPLOADS_TUS, required"`
	LocalFolderUploadsA   string `env:"LOCAL_FOLDER_UPLOADS_A, required"`

	// TUSD
	TusdHandlerBasePath string `env:"TUSD_HANDLER_BASE_PATH, required"`

	// Processing Status
	ProcessingStatusURI string `env:"PROCESSING_STATUS_URI, required"`

	// Azure TUS Upload storage
	TusAzStorageConfig *AzureStorageConfig `env:", prefix=TUS_"`
	// DexAzStorageConfig *AzureStorageConfig `env:", prefix="DEX_"` this is currently same as TUS above only different container name
	DexAzStorageContainerName string `env:"DEX_AZ_STORAGE_CONTAINER_NAME"`
	//
	RouterAzStorageConfig *AzureStorageConfig `env:", prefix=ROUTER_"`
	EdavAzStorageConfig   *AzureStorageConfig `env:", prefix=EDAV_"`
} // .AppConfig

type AzureStorageConfig struct {
	AzStorageName         string `env:"AZ_STORAGE_NAME"`
	AzStorageKey          string `env:"AZ_STORAGE_KEY"`
	AzContainerName       string `env:"AZ_CONTAINER_NAME"`
	AzContainerEndpoint   string `env:"AZ_CONTAINER_ENDPOINT"`
	AzContainerAccessType string `env:"AZ_CONTAINER_ACCESS_TYPE"`
} // .AzureStorageConfig

// ParseConfig loads app configuration based on environment variables and returns AppConfig struct
func ParseConfig(ctx context.Context) (AppConfig, error) {

	var ac AppConfig
	if err := envconfig.Process(ctx, &ac); err != nil {
		return AppConfig{}, err
	} // .if

	return ac, nil
} // .ParseConfig
