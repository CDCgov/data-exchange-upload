package appconfig

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

type AppConfig struct {

	// App and for Logger
	System     string `env:"SYSTEM, required"`
	DexProduct string `env:"DEX_PRODUCT, required"`
	DexApp     string `env:"DEX_APP, required"`
	LoggerDebugOn bool `env:"LOGGER_DEBUG_ON"`

	// Server
	ServerPort string `env:"SERVER_PORT, required"`

	// Metadata
	MetadataVersions string `env:"METADATA_VERSIONS, required"`

	// Metadata v1
	AllowedDestAndEventsPath string `env:"ALLOWED_DEST_AND_EVENTS_PATH, required"`
	DefinitionsPath          string `env:"DEFINITIONS_PATH, required"`
	UploadConfigPath         string `env:"UPLOAD_CONFIG_PATH, required"`

	// Azure
	AzStorage                        string `env:"AZ_STORAGE"`
	AzContainerAccessType            string `env:"AZ_CONTAINER_ACCESS_TYPE"`
	AzBlobAccessTier                 string `env:"AZ_BLOB_ACCESS_TIER"`
	AzObjectPrefix                   string `env:"AZ_OBJECT_PREFIX"`
	AzEndpoint                       string `env:"AZ_ENDPOINT"`

} // .AppConfig

func ParseConfig() (AppConfig, error) { 

	ctx := context.Background()

	var ac AppConfig
	if err := envconfig.Process(ctx, &ac); err != nil {
	  return AppConfig{}, err 
	} // .if

	return ac, nil 
} // .ParseConfig
