package appconfig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
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
	// Common Configs

	// Logging and Environment
	LoggerDebugOn bool `env:"LOGGER_DEBUG_ON"`
	//QUESTION: this is arbitrary so is it useful?
	Environment string `env:"ENVIRONMENT, default=DEV"`

	// Server Configs
	ServerProtocol                string `env:"SERVER_PROTOCOL, default=http"`
	ServerHostname                string `env:"SERVER_HOSTNAME, default=localhost"`
	ServerPort                    string `env:"SERVER_PORT, default=8080"`
	TusdHandlerBasePath           string `env:"TUSD_HANDLER_BASE_PATH, default=/files/"`
	TusdHandlerInfoPath           string `env:"TUSD_HANDLER_INFO_PATH, default=/info/"`
	EventMaxRetryCount            int    `env:"EVENT_MAX_RETRY_COUNT, default=3"`
	ExternalServerUrl             string
	InternalServerUrl             string
	InternalServerFileEndpointUrl string
	InternalServerInfoEndpointUrl string
	ExternalServerFileEndpointUrl string
	ExternalServerInfoEndpointUrl string
	Metrics                       MetricsConfig `env:", prefix=METRICS_"`

	// TUSD
	TusUploadPrefix string `env:"TUS_UPLOAD_PREFIX, default=tus-prefix"`

	// User Interface Configs
	UIPort                   string `env:"UI_PORT, default=8081"`
	UIServerExternalProtocol string `env:"UI_SERVER_EXTERNAL_PROTOCOL, default=http"`
	UIServerInternalProtocol string `env:"UI_SERVER_INTERNAL_PROTOCOL, default=http"`
	UIServerExternalHost     string `env:"UI_SERVER_EXTERNAL_HOST, default=localhost:8080"`
	UIServerInternalHost     string `env:"UI_SERVER_INTERNAL_HOST, default=localhost:8080"`
	// CsrfToken                string `env:"CSRF_TOKEN, default=1qQBJumxRABFBLvaz5PSXBcXLE84viE42x4Aev359DvLSvzjbXSme3whhFkESatW"`
	CSRF *CSRFConfig `env:", prefix=CSRF_"`
	// WARNING: the default CsrfToken value is for local development use only, it needs to be replaced by a secret 32 byte string before being used in production

	// TUS Upload file lock
	TusRedisLockURI string `env:"REDIS_CONNECTION_STRING"`

	// OAuth Configs
	OauthConfig *OauthConfig `env:", prefix=OAUTH_"`

	// process status health
	ProcessingStatusHealthURI string `env:"PROCESSING_STATUS_HEALTH_URI"`

	// Local File System Configs
	LocalFolderUploadsTus string `env:"LOCAL_FOLDER_UPLOADS_TUS, default=./uploads"`
	UploadConfigPath      string `env:"UPLOAD_CONFIG_PATH, default=../upload-configs"`
	// Local File System Report Directory
	LocalReportsFolder string `env:"LOCAL_REPORTS_FOLDER, default=./uploads/reports"`
	LocalEventsFolder  string `env:"LOCAL_EVENTS_FOLDER, default=./uploads/events"`

	// Azure Storage Configs
	AzureConnection              *AzureStorageConfig `env:", prefix=AZURE_, noinit"`
	AzureUploadContainer         string              `env:"TUS_AZURE_CONTAINER_NAME"`
	AzureManifestConfigContainer string              `env:"DEX_MANIFEST_CONFIG_CONTAINER_NAME"`

	// Azure Report Queue
	ReporterConnection *AzureQueueConfig `env:", prefix=REPORTER_, noinit"`
	// Azure Event Topics
	// Azure Event Publisher Topic
	PublisherConnection *AzureQueueConfig `env:", prefix=PUBLISHER_,noinit"`
	// Azure Event Subscriber Subscription
	SubscriberConnection *AzureQueueConfig `env:", prefix=SUBSCRIBER_,noinit"`

	SNSReporterConnection   *SNSConfig `env:", prefix=SNS_REPORTER_,noinit"`
	SNSPublisherConnection  *SNSConfig `env:", prefix=SNS_PUBLISHER_,noinit"`
	SQSSubscriberConnection *SQSConfig `env:", prefix=SQS_SUBSCRIBER_,noinit"`

	// S3 Storage Configs
	S3Connection           *S3StorageConfig `env:", prefix=S3_, noinit"`
	S3ManifestConfigBucket string           `env:"DEX_MANIFEST_CONFIG_BUCKET_NAME"`
	S3ManifestConfigFolder string           `env:"DEX_S3_MANIFEST_CONFIG_FOLDER_NAME"`

	DeliveryConfigFile string `env:"DEX_DELIVERY_CONFIG_FILE, default=./configs/local/deliver.yml"`
	ListenerWorkers    int    `env:"LISTENER_WORKERS, default=5"`
} // .AppConfig

type MetricsConfig struct {
	LabelsFromManifest []string `env:"LABELS_FROM_MANIFEST, default=data_stream_id,data_stream_route,sender_id"`
	PollIntervalMillis int      `env:"POLL_INTERVAL_MILLIS, default=3000"`
}

func (conf *AppConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(RootResp{
		System:     "DEX",
		DexProduct: "UPLOAD API",
		DexApp:     "upload server",
		ServerTime: time.Now().Format(time.RFC3339Nano),
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

type SNSConfig struct {
	EventArn string `env:"EVENT_ARN"`
}

type SQSConfig struct {
	EventArn    string `env:"EVENT_ARN"`
	MaxMessages int    `env:"MAX_MESSAGES"`
	TopicArn    string `env:"TOPIC_ARN"`
	MaxRetries  int    `env:"MAX_RETRIES"`
}

type AzureStorageConfig struct {
	StorageName       string `env:"STORAGE_ACCOUNT"`
	StorageKey        string `env:"STORAGE_KEY"`
	TenantId          string `env:"TENANT_ID"`
	ClientId          string `env:"CLIENT_ID"`
	ClientSecret      string `env:"CLIENT_SECRET"`
	ContainerEndpoint string `env:"ENDPOINT"`
} // .AzureStorageConfig

func (ac *AzureStorageConfig) Credentials() storeaz.Credentials {
	return storeaz.Credentials{
		StorageName:       ac.StorageName,
		StorageKey:        ac.StorageKey,
		TenantId:          ac.TenantId,
		ClientId:          ac.ClientId,
		ClientSecret:      ac.ClientSecret,
		ContainerEndpoint: ac.ContainerEndpoint,
	}
}

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
	MaxMessages      int    `env:"MAX_MESSAGES"`
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
	SessionKey       string `env:"SESSION_KEY"`
	SessionSecure    bool   `env:"SESSION_SECURE, default=true"`
	SessionDomain    string `env:"SESSION_DOMAIN"`
}

type CSRFConfig struct {
	Token          string `env:"TOKEN, default=1qQBJumxRABFBLvaz5PSXBcXLE84viE42x4Aev359DvLSvzjbXSme3whhFkESatW"`
	TrustedOrigins string `env:"TRUSTED_ORIGINS, default=localhost:8081"`
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
			return AppConfig{}, fmt.Errorf("missing required values for connecting to Azure")
		}
		if ac.AzureConnection.ContainerEndpoint == "" {
			ac.AzureConnection.ContainerEndpoint = fmt.Sprintf("https://%s.blob.core.windows.net", ac.AzureConnection.StorageName)
		}
	}

	if ac.S3Connection != nil {
		if ac.S3Connection.BucketName == "" || ac.S3Connection.Endpoint == "" {
			return AppConfig{}, fmt.Errorf("missing required values for connecting to AWS S3")
		}
	}

	ac.InternalServerUrl = fmt.Sprintf("%s://%s", ac.UIServerInternalProtocol, ac.UIServerInternalHost)
	ac.ExternalServerUrl = fmt.Sprintf("%s://%s", ac.UIServerExternalProtocol, ac.UIServerExternalHost)
	ac.ExternalServerFileEndpointUrl = ac.ExternalServerUrl + ac.TusdHandlerBasePath
	ac.ExternalServerInfoEndpointUrl = ac.ExternalServerUrl + ac.TusdHandlerInfoPath
	ac.InternalServerFileEndpointUrl = ac.InternalServerUrl + ac.TusdHandlerBasePath
	ac.InternalServerInfoEndpointUrl = ac.InternalServerUrl + ac.TusdHandlerInfoPath

	LoadedConfig = &ac
	return ac, nil
} // .ParseConfig
