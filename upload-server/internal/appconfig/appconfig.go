package appconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
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
	Environment        string `env:"ENVIRONMENT, default=DEV"`
	EventMaxRetryCount int    `env:"EVENT_MAX_RETRY_COUNT, default=3"`

	UploadConfigPath string `env:"UPLOAD_CONFIG_PATH, default=../upload-configs"`

	LocalFolderUploadsTus string `env:"LOCAL_FOLDER_UPLOADS_TUS, default=./uploads"`
	LocalReportsFolder    string `env:"LOCAL_REPORTS_FOLDER, default=./uploads/reports"`
	LocalEventsFolder     string `env:"LOCAL_EVENTS_FOLDER, default=./uploads/events"`
	LocalDEXFolder        string `env:"LOCAL_DEX_FOLDER, default=./uploads/dex"`
	LocalEDAVFolder       string `env:"LOCAL_EDAV_FOLDER, default=./uploads/edav"`
	LocalRoutingFolder    string `env:"LOCAL_ROUTING_FOLDER, default=./uploads/routing"`

	// OAUTH
	OauthIssuerUrl      string `env:"OAUTH_ISSUER_URL, default=https://apigw-stg.cdc.gov:8443"`
	OauthRequiredScopes string `env:"OAUTH_REQUIRED_SCOPES"`

	// TUSD
	TusdHandlerBasePath string `env:"TUSD_HANDLER_BASE_PATH, default=/files/"`

	// UI
	TusUIFileEndpointUrl string `env:"TUS_UI_FILE_ENDPOINT_URL, default=http://localhost:8080/files/"`
	TusUIInfoEndpointUrl string `env:"TUS_UI_INFO_ENDPOINT_URL, default=http://localhost:8080/info/"`
	UIPort               string `env:"UI_PORT, default=:8081"`

	// Processing Status
	ProcessingStatusHealthURI string `env:"PROCESSING_STATUS_HEALTH_URI"`

	AzureConnection      *AzureStorageConfig `env:", prefix=AZURE_, noinit"`
	S3Connection         *S3StorageConfig    `env:", prefix=AWS_, noinit"`
	EdavConnection       *AzureStorageConfig `env:", prefix=EDAV_, noinit"`
	RoutingConnection    *AzureStorageConfig `env:", prefix=ROUTING_, noinit"`
	PublisherConnection  *AzureQueueConfig   `env:", prefix=PUBLISHER_,noinit"`
	SubscriberConnection *AzureQueueConfig   `env:", prefix=SUBSCRIBER_,noinit"`

	// Reporting
	ReporterConnection *AzureQueueConfig `env:", prefix=REPORTER_, noinit"`

	// Azure TUS Upload storage
	TusRedisLockURI              string `env:"REDIS_CONNECTION_STRING"`
	AzureUploadContainer         string `env:"TUS_AZURE_CONTAINER_NAME"`
	AzureManifestConfigContainer string `env:"DEX_MANIFEST_CONFIG_CONTAINER_NAME"`
	TusUploadPrefix              string `env:"TUS_UPLOAD_PREFIX, default=tus-prefix"`

	// S3
	S3ManifestConfigBucket string `env:"DEX_MANIFEST_CONFIG_BUCKET_NAME"`

	// Upload processing
	DexCheckpointContainer     string `env:"DEX_CHECKPOINT_CONTAINER_NAME, default=dex-checkpoint"`
	EdavCheckpointContainer    string `env:"EDAV_CHECKPOINT_CONTAINER_NAME, default=edav-checkpoint"`
	RoutingCheckpointContainer string `env:"ROUTING_CHECKPOINT_CONTAINER_NAME, default=routing-checkpoint"`
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

type S3StorageConfig struct {
	Endpoint   string `env:"S3_ENDPOINT"`
	BucketName string `env:"S3_BUCKET_NAME"`
}

type AzureContainerConfig struct {
	AzureStorageConfig
	ContainerName string
}

type AzureQueueConfig struct {
	ConnectionString string `env:"CONNECTION_STRING"`
	Topic            string `env:"TOPIC"`
	Queue            string `env:"QUEUE"`
	Subscription     string `env:"SUBSCRIPTION"`
}

type LocalStorageConfig struct {
	FromPathStr string
	FromPath    fs.FS
	ToPath      string
}

func (azc *AzureStorageConfig) Check() error {
	errs := []error{}
	if azc.StorageName == "" {
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzStorageName",
		})
	}
	if azc.StorageKey == "" {
		// TODO check in using service principle
		errs = append(errs, &MissingConfigError{
			ConfigName: "AzStorageKey",
		})
	}
	return errors.Join(errs...)
}

func GetAzureContainerConfig(target string) (*AzureContainerConfig, error) {
	switch target {
	case "dex":
		return &AzureContainerConfig{
			AzureStorageConfig: *LoadedConfig.AzureConnection,
			ContainerName:      LoadedConfig.DexCheckpointContainer,
		}, nil
	case "edav":
		return &AzureContainerConfig{
			AzureStorageConfig: *LoadedConfig.EdavConnection,
			ContainerName:      LoadedConfig.EdavCheckpointContainer,
		}, nil
	case "routing":
		return &AzureContainerConfig{
			AzureStorageConfig: *LoadedConfig.RoutingConnection,
			ContainerName:      LoadedConfig.RoutingCheckpointContainer,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported azure target %s", target)
	}
}

func LocalStoreConfig(target string, appConfig *AppConfig) (*LocalStorageConfig, error) {
	fromPathStr := appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix
	fromPath := os.DirFS(fromPathStr)

	switch target {
	case "dex":
		return &LocalStorageConfig{
			FromPathStr: fromPathStr,
			FromPath:    fromPath,
			ToPath:      appConfig.LocalDEXFolder,
		}, nil
	case "edav":
		return &LocalStorageConfig{
			FromPathStr: fromPathStr,
			FromPath:    fromPath,
			ToPath:      appConfig.LocalEDAVFolder,
		}, nil
	case "routing":
		return &LocalStorageConfig{
			FromPathStr: fromPathStr,
			FromPath:    fromPath,
			ToPath:      appConfig.LocalRoutingFolder,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported local target %s", target)
	}
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
