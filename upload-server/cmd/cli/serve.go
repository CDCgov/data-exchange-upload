package cli

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/redislockerhealth"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/sbhealth"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/redislocker"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/memorylocker"
	"net/http"
	"strings"
)

func Serve(ctx context.Context, appConfig appconfig.AppConfig) (http.Handler, error) {
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
	err = InitReporters(ctx, appConfig)
	if err != nil {
		return nil, err
	}
	defer reports.DefaultReporter.Close()

	var fileReadyPublisher event.Publisher
	fileReadyPublisher = &event.MemoryPublisher{
		Dir: appConfig.LocalEventsFolder,
	}

	if appConfig.PublisherConnection != nil {
		client, err := event.NewAMQPServiceBusClient(appConfig.PublisherConnection.ConnectionString)
		if err != nil {
			logger.Error("failed to connect to event service bus", "error", err)
			return nil, err
		}
		sender, err := client.NewSender(appConfig.PublisherConnection.Topic, nil)
		if err != nil {
			logger.Error("failed to configure event publisher", "error", err)
			return nil, err
		}
		adminClient, err := admin.NewClientFromConnectionString(appConfig.PublisherConnection.ConnectionString, nil)
		if err != nil {
			logger.Error("failed to connect to service bus admin client", "error", err)
			return nil, err
		}

		fileReadyPublisher = &event.AzurePublisher{
			Context:     ctx,
			EventType:   event.FileReadyEventType,
			Sender:      sender,
			Config:      *appConfig.PublisherConnection,
			AdminClient: adminClient,
		}
		defer fileReadyPublisher.Close()

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
	http.Handle("/version", &VersionHandler{})

	return http.DefaultServeMux, nil
}
