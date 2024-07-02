package cli

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/aznamespaces"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/redislockerhealth"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/sbhealth"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/redislocker"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/memorylocker"
	"net/http"
	"strings"
)

func Serve(ctx context.Context, appConfig appconfig.AppConfig, fileReadyChan chan event.FileReadyEvent) (http.Handler, error) {
	if sloger.DefaultLogger != nil {
		logger = sloger.DefaultLogger
	}
	// initialize processing status health checker
	sbHealth, err := sbhealth.New(appConfig)
	if err != nil {
		logger.Error("error initializing service bus health check", "error", err)
	}
	if sbHealth != nil {
		health.Register(sbHealth)
	}

	// create and register data store
	store, storeHealthCheck, err := GetDataStore(appConfig)
	if err != nil {
		logger.Error("error starting app, error configuring storage", "error", err)
		return nil, err
	} // .if
	health.Register(storeHealthCheck)

	uploadInfoHandler, err := GetUploadInfoHandler(&appConfig)
	if err != nil {
		logger.Error("error configuring upload info handler: ", "error", err)
		return nil, err
	}

	// initialize locker
	var locker handlertusd.Locker = memorylocker.New()
	if appConfig.TusRedisLockURI != "" {
		var err error
		locker, err = redislocker.New(appConfig.TusRedisLockURI, redislocker.WithLogger(logger))
		if err != nil {
			logger.Error("failed to configure Redis locker, defaulting to in-memory locker", "error", err)
		}

		// configure redislocker health check
		redisLockerHealth, err := redislockerhealth.New(appConfig.TusRedisLockURI)
		if err != nil {
			logger.Error("failed to configure Redis locker health check, skipping check", "error", err)
		} else {
			health.Register(redisLockerHealth)
		}
	}

	// initialize event reporter
	err = InitReporters(appConfig)

	var fileReadyPublisher event.Publisher
	fileReadyPublisher = &event.MemoryPublisher{
		FileReadyChannel: fileReadyChan,
		Dir:              appConfig.LocalEventsFolder,
	}

	if appConfig.QueueConnection != nil {
		cred := azcore.NewKeyCredential(appConfig.QueueConnection.AccessKey)
		client, err := aznamespaces.NewSenderClientWithSharedKeyCredential(appConfig.QueueConnection.Endpoint, appConfig.QueueConnection.Topic, cred, nil)
		if err != nil {
			logger.Error("failed to connect to azure event grid")
		}

		fileReadyPublisher = &event.AzurePublisher{
			Client: client,
			Config: *appConfig.QueueConnection,
		}

		health.Register(fileReadyPublisher)
	}

	// get and initialize tusd hook handlers
	hookHandler, err := GetHookHandler(ctx, appConfig, fileReadyPublisher)
	if err != nil {
		logger.Error("error configuring tusd handler: ", "error", err)
		return nil, err
	}

	// initialize tusd handler
	handlerTusd, err := handlertusd.New(store, locker, hookHandler, appConfig.TusdHandlerBasePath)
	if err != nil {
		logger.Error("error starting tusd handler: ", "error", err)
		return nil, err
	}

	// --------------------------------------------------------------
	// 	TUSD handler
	// --------------------------------------------------------------
	// Route for TUSD to start listening on and accept http request
	logger.Info("hosting tus handler", "path", appConfig.TusdHandlerBasePath)
	pathWithoutSlash := strings.TrimSuffix(appConfig.TusdHandlerBasePath, "/")
	pathWithSlash := pathWithoutSlash + "/"
	http.Handle(pathWithoutSlash, http.StripPrefix(pathWithoutSlash, handlerTusd))
	http.Handle(pathWithSlash, http.StripPrefix(pathWithSlash, handlerTusd))

	// initialize and route handler for DEX
	handlerDex := handlerdex.New(appConfig)
	http.Handle("/", handlerDex)

	// --------------------------------------------------------------
	// 	Prometheus metrics handler for /metrics
	// --------------------------------------------------------------
	hooks.SetupHookMetrics()
	http.Handle("/metrics", promhttp.Handler())
	setupMetrics()

	http.Handle("/info/{UploadID}", uploadInfoHandler)

	return http.DefaultServeMux, nil
}
