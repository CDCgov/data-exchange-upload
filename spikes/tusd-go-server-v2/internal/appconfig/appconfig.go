package appconfig

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/sethvargo/go-envconfig"
) // .import

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

type RootResp struct {
	System     string `json:"system"`
	DexProduct string `json:"dex_product"`
	DexApp     string `json:"dex_app"`
	ServerTime string `json:"server_time"`
} // .rootResp

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
	HydrateV1ConfigPath      string `env:"HYDRATE_V1_CONFIG_PATH, required"`

	// Local folder path e.g. ../uploads
	LocalFolderUploadsTus string `env:"LOCAL_FOLDER_UPLOADS_TUS, required"`
	LocalFolderUploadsA   string `env:"LOCAL_FOLDER_UPLOADS_A, required"`

	// TUSD
	TusdHandlerBasePath string `env:"TUSD_HANDLER_BASE_PATH, required"`

	// Processing Status
	ProcessingStatusHealthURI           string `env:"PROCESSING_STATUS_HEALTH_URI, required"`
	ProcessingStatusServiceBusNamespace string `env:"PROCESSING_STATUS_SERVICE_BUS_NAMESPACE, required"`
	ProcessingStatusServiceBusQueue     string `env:"PROCESSING_STATUS_SERVICE_BUS_QUEUE"`

	// Azure TUS Upload storage
	TusAzStorageConfig *AzureStorageConfig `env:", prefix=TUS_"`
	// DexAzStorageConfig *AzureStorageConfig `env:", prefix="DEX_"` this is currently same as TUS above only different container name
	DexAzStorageContainerName string `env:"DEX_AZ_STORAGE_CONTAINER_NAME"`
	//
	RouterAzStorageConfig *AzureStorageConfig `env:", prefix=ROUTER_"`
	EdavAzStorageConfig   *AzureStorageConfig `env:", prefix=EDAV_"`

	CopyRetryTimes int `env:"COPY_RETRY_TIMES, required"`
	CopyRetryDelay int `env:"COPY_RETRY_DELAY, required"`
} // .AppConfig

func (conf *AppConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(RootResp{
		System:     conf.System,
		DexProduct: conf.DexProduct,
		DexApp:     conf.DexApp,
		ServerTime: time.Now().Format(time.RFC3339),
	}) // .jsonResp
	if err != nil {
		errMsg := "error marshal json for root response"
		logger.Error(errMsg, "error", err.Error())
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	} // .if

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

type AzureStorageConfig struct {
	AzStorageName         string `env:"AZ_STORAGE_NAME"`
	AzStorageKey          string `env:"AZ_STORAGE_KEY"`
	AzContainerName       string `env:"AZ_CONTAINER_NAME"`
	AzContainerEndpoint   string `env:"AZ_CONTAINER_ENDPOINT"`
	AzContainerAccessType string `env:"AZ_CONTAINER_ACCESS_TYPE"`
} // .AzureStorageConfig

func (azc *AzureStorageConfig) Check() error {
	errs := []error{}
	if azc.AzStorageName == "" {
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzStorageName",
		})
	}
	if azc.AzStorageName == "" {
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzStorageKey",
		})
	}
	if azc.AzStorageName == "" {
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzContainerName",
		})
	}
	if azc.AzStorageName == "" {
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzContainerEndpoint",
		})
	}
	if azc.AzStorageName == "" {
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzContainerAccessType",
		})
	}
	return errors.Join(errs...)
}

var LoadedConfig = &AppConfig{}

func Handler() http.Handler {
	return LoadedConfig
}

// ParseConfig loads app configuration based on environment variables and returns AppConfig struct
func ParseConfig(ctx context.Context) (AppConfig, error) {

	var ac AppConfig
	if err := envconfig.Process(ctx, &ac); err != nil {
		return AppConfig{}, err
	} // .if
	LoadedConfig = &ac
	return ac, nil
} // .ParseConfig
