package appconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
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
	LoggerDebugOn bool `env:"LOGGER_DEBUG_ON"`

	// Server
	ServerPort string `env:"SERVER_PORT, default=8080"`
	//QUESTION: this is arbitrary so is it useful?
	Environment string `env:"ENVIRONMENT, default=DEV"`

	UploadConfigPath string `env:"UPLOAD_CONFIG_PATH, default=../upload-configs"`

	LocalFolderUploadsTus string `env:"LOCAL_FOLDER_UPLOADS_TUS, default=./uploads"`
	LocalReportsFolder    string `env:"LOCAL_REPORTS_FOLDER, default=./uploads/reports"`

	// TUSD
	TusdHandlerBasePath string `env:"TUSD_HANDLER_BASE_PATH, default=/files/"`

	// Processing Status
	ProcessingStatusHealthURI           string `env:"PROCESSING_STATUS_HEALTH_URI"`
	ProcessingStatusServiceBusNamespace string `env:"PROCESSING_STATUS_SERVICE_BUS_NAMESPACE"`
	ProcessingStatusServiceBusQueue     string `env:"PROCESSING_STATUS_SERVICE_BUS_QUEUE"`

	AzureConnection            *AzureStorageConfig `env:", prefix=AZURE_, noinit"`
	ServiceBusConnectionString string              `env:"SERVICE_BUS_CONNECTION_STR"`
	ReportQueueName            string              `env:"REPORT_QUEUE_NAME, default=processing-status-cosmos-db-queue"`

	// Azure TUS Upload storage
	TusRedisLockURI              string `env:"REDIS_CONNECTION_STRING"`
	AzureUploadContainer         string `env:"TUS_AZURE_CONTAINER_NAME"`
	AzureManifestConfigContainer string `env:"DEX_MANIFEST_CONFIG_CONTAINER_NAME"`
	TusUploadPrefix              string `env:"TUS_UPLOAD_PREFIX, default=tus-prefix"`
} // .AppConfig

func (conf *AppConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(RootResp{
		System:     "DEX",
		DexProduct: "UPLOAD API",
		DexApp:     "upload server",
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
	StorageName       string `env:"STORAGE_ACCOUNT"`
	StorageKey        string `env:"STORAGE_KEY"`
	ContainerEndpoint string `env:"ENDPOINT"`
} // .AzureStorageConfig

func (azc *AzureStorageConfig) Check() error {
	errs := []error{}
	if azc.StorageName == "" {
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzStorageName",
		})
	}
	if azc.StorageKey == "" {
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzStorageKey",
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
	if ac.AzureConnection != nil {
		if ac.AzureConnection.ContainerEndpoint == "" {
			ac.AzureConnection.ContainerEndpoint = fmt.Sprintf("https://%s.blob.core.windows.net", ac.AzureConnection.StorageName)
		}
	}
	LoadedConfig = &ac
	return ac, nil
} // .ParseConfig
