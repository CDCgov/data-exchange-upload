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
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/sethvargo/go-envconfig"
) // .import

var DeliveryTargetEdav = "edav"
var DeliveryTargetEhdi = "ehdi"
var DeliveryTargetEicr = "eicr"
var DeliveryTargetNcird = "ncird"
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

	//QUESTION: this is arbitrary so is it useful?
	Environment string `env:"ENVIRONMENT, default=DEV"`

	// Server
	ServerProtocol        string `env:"SERVER_PROTOCOL, default=http"`
	ServerHostname        string `env:"SERVER_HOSTNAME, default=localhost"`
	ServerPort            string `env:"SERVER_PORT, default=8080"`
	TusdHandlerBasePath   string `env:"TUSD_HANDLER_BASE_PATH, default=/files/"`
	TusdHandlerInfoPath   string `env:"TUSD_HANDLER_INFO_PATH, default=/info/"`
	UploadConfigPath      string `env:"UPLOAD_CONFIG_PATH, default=../upload-configs"`
	EventMaxRetryCount    int    `env:"EVENT_MAX_RETRY_COUNT, default=3"`
	ServerUrl             string
	ServerFileEndpointUrl string
	ServerInfoEndpointUrl string

	// process status health
	ProcessingStatusHealthURI string `env:"PROCESSING_STATUS_HEALTH_URI"`

	// metrics
	Metrics MetricsConfig `env:", prefix=METRICS_"`

	// TUSD
	TusUploadPrefix string `env:"TUS_UPLOAD_PREFIX, default=tus-prefix"`
	// TUS Upload file lock
	TusRedisLockURI string `env:"REDIS_CONNECTION_STRING"`

	// oauth
	OauthConfig *OauthConfig `env:", prefix=OAUTH_"`

	// UI
	UIPort    string `env:"UI_PORT, default=:8081"`
	CsrfToken string `env:"CSRF_TOKEN, required"`
	// WARNING: the default CsrfToken value is for local development use only, it needs to be replaced by a secret 32 byte string before being used in production

	// Local file system upload config
	LocalFolderUploadsTus string `env:"LOCAL_FOLDER_UPLOADS_TUS, default=./uploads"`
	LocalReportsFolder    string `env:"LOCAL_REPORTS_FOLDER, default=./uploads/reports"`
	LocalEventsFolder     string `env:"LOCAL_EVENTS_FOLDER, default=./uploads/events"`

	// Local file system delivery config
	LocalRoutingFolder string `env:"LOCAL_ROUTING_FOLDER, default=./uploads/routing"`
	LocalEDAVFolder    string `env:"LOCAL_EDAV_FOLDER, default=./uploads/edav"`
	LocalEhdiFolder    string `env:"LOCAL_EHDI_FOLDER, default=./uploads/ehdi"`
	LocalEicrFolder    string `env:"LOCAL_EICR_FOLDER, default=./uploads/eicr"`
	LocalNcirdFolder   string `env:"LOCAL_NCIRD_FOLDER, default=./uploads/ncird"`

	// Azure upload config
	AzureConnection              *AzureStorageConfig `env:", prefix=AZURE_, noinit"`
	AzureUploadContainer         string              `env:"TUS_AZURE_CONTAINER_NAME"`
	AzureManifestConfigContainer string              `env:"DEX_MANIFEST_CONFIG_CONTAINER_NAME"`

	ReporterConnection   *AzureQueueConfig `env:", prefix=REPORTER_, noinit"`
	PublisherConnection  *AzureQueueConfig `env:", prefix=PUBLISHER_,noinit"`
	SubscriberConnection *AzureQueueConfig `env:", prefix=SUBSCRIBER_,noinit"`

	// Azure delivery config
	RoutingConnection          *AzureStorageConfig `env:", prefix=ROUTING_, noinit"`
	RoutingCheckpointContainer string              `env:"ROUTING_CHECKPOINT_CONTAINER_NAME, default=routing-checkpoint"`

	EdavConnection          *AzureStorageConfig `env:", prefix=EDAV_, noinit"`
	EdavCheckpointContainer string              `env:"EDAV_CHECKPOINT_CONTAINER_NAME, default=edav-checkpoint"`

	EhdiConnection          *AzureStorageConfig `env:", prefix=EHDI_, noinit"`
	EhdiCheckpointContainer string              `env:"EHDI_CHECKPOINT_CONTAINER_NAME, default=ehdi-checkpoint"`

	EicrConnection          *AzureStorageConfig `env:", prefix=EICR_, noinit"`
	EicrCheckpointContainer string              `env:"EICR_CHECKPOINT_CONTAINER_NAME, default=eicr-checkpoint"`

	NcirdConnection          *AzureStorageConfig `env:", prefix=NCIRD_, noinit"`
	NcirdCheckpointContainer string              `env:"NCIRD_CHECKPOINT_CONTAINER_NAME, default=ncird-checkpoint"`

	// S3 upload config
	S3Connection           *S3StorageConfig `env:", prefix=S3_, noinit"`
	S3ManifestConfigBucket string           `env:"DEX_MANIFEST_CONFIG_BUCKET_NAME"`
	S3ManifestConfigFolder string           `env:"DEX_S3_MANIFEST_CONFIG_FOLDER_NAME"`

	// S3 delivery config
	RoutingS3Connection *S3StorageConfig `env:", prefix=ROUTING_S3_, noinit"`
	EdavS3Connection    *S3StorageConfig `env:", prefix=EDAV_S3_, noinit"`
	EhdiS3Connection    *S3StorageConfig `env:", prefix=EHDI_S3_, noinit"`
	EicrS3Connection    *S3StorageConfig `env:", prefix=EICR_S3_, noinit"`
	NcirdS3Connection   *S3StorageConfig `env:", prefix=NCIRD_S3_, noinit"`
} // .AppConfig

type MetricsConfig struct {
	LabelsFromManifest []string `env:"LABELS_FROM_MANIFEST, default=data_stream_id,data_stream_route,sender_id"`
}

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
	TenantId          string `env:"TENANT_ID"`
	ClientId          string `env:"CLIENT_ID"`
	ClientSecret      string `env:"CLIENT_SECRET"`
	ContainerEndpoint string `env:"ENDPOINT"`
} // .AzureStorageConfig

type S3StorageConfig struct {
	Endpoint   string `env:"ENDPOINT"`
	BucketName string `env:"BUCKET_NAME"`
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

type OauthConfig struct {
	AuthEnabled      bool   `env:"AUTH_ENABLED, default=false"`
	IntrospectionUrl string `env:"INTROSPECTION_URL"`
	IssuerUrl        string `env:"ISSUER_URL"`
	RequiredScopes   string `env:"REQUIRED_SCOPES"`
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
	case DeliveryTargetEdav:
		return &AzureContainerConfig{
			AzureStorageConfig: *LoadedConfig.EdavConnection,
			ContainerName:      LoadedConfig.EdavCheckpointContainer,
		}, nil
	case "routing":
		return &AzureContainerConfig{
			AzureStorageConfig: *LoadedConfig.RoutingConnection,
			ContainerName:      LoadedConfig.RoutingCheckpointContainer,
		}, nil
	case DeliveryTargetEhdi:
		return &AzureContainerConfig{
			AzureStorageConfig: *LoadedConfig.EhdiConnection,
			ContainerName:      LoadedConfig.EhdiCheckpointContainer,
		}, nil
	case DeliveryTargetEicr:
		return &AzureContainerConfig{
			AzureStorageConfig: *LoadedConfig.EicrConnection,
			ContainerName:      LoadedConfig.EicrCheckpointContainer,
		}, nil
	case DeliveryTargetNcird:
		return &AzureContainerConfig{
			AzureStorageConfig: *LoadedConfig.NcirdConnection,
			ContainerName:      LoadedConfig.NcirdCheckpointContainer,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported azure target %s", target)
	}
}

func LocalStoreConfig(target string, appConfig *AppConfig) (*LocalStorageConfig, error) {
	fromPathStr := filepath.Join(appConfig.LocalFolderUploadsTus, appConfig.TusUploadPrefix)
	fromPath := os.DirFS(fromPathStr)

	switch target {
	case DeliveryTargetEdav:
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
	case DeliveryTargetEhdi:
		return &LocalStorageConfig{
			FromPathStr: fromPathStr,
			FromPath:    fromPath,
			ToPath:      appConfig.LocalEhdiFolder,
		}, nil
	case DeliveryTargetEicr:
		return &LocalStorageConfig{
			FromPathStr: fromPathStr,
			FromPath:    fromPath,
			ToPath:      appConfig.LocalEicrFolder,
		}, nil
	case DeliveryTargetNcird:
		return &LocalStorageConfig{
			FromPathStr: fromPathStr,
			FromPath:    fromPath,
			ToPath:      appConfig.LocalNcirdFolder,
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
		if ac.AzureConnection.StorageName == "" || ac.AzureConnection.StorageKey == "" {
			ac.AzureConnection = nil
		} else if ac.AzureConnection.ContainerEndpoint == "" {
			ac.AzureConnection.ContainerEndpoint = fmt.Sprintf("https://%s.blob.core.windows.net", ac.AzureConnection.StorageName)
		}
	}

	if ac.S3Connection != nil {
		if ac.S3Connection.BucketName == "" || ac.S3Connection.Endpoint == "" {
			ac.S3Connection = nil
		}
	}

	ac.ServerUrl = fmt.Sprintf("%s://%s:%s", ac.ServerProtocol, ac.ServerHostname, ac.ServerPort)
	ac.ServerFileEndpointUrl = ac.ServerUrl + ac.TusdHandlerBasePath
	ac.ServerInfoEndpointUrl = ac.ServerUrl + ac.TusdHandlerInfoPath

	LoadedConfig = &ac
	return ac, nil
} // .ParseConfig
