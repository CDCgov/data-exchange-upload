package cli

import (
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/processingstatus"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/redislocker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/memorylocker"
)

func Serve(appConfig appconfig.AppConfig) (http.Handler, error) {

	// initialize processing status sender
	psSender, err := processingstatus.New(appConfig)
	if err != nil {
		logger.Error("error initializing processing status not available", "error", err)
	}
	if psSender != nil {
		health.Register(psSender)
	}

	// create and register data store
	store, storeHealthCheck, err := CreateDataStore(appConfig)
	if err != nil {
		logger.Error("error starting app, error configuring storage", "error", err)
		return nil, err
	}
	health.Register(storeHealthCheck)

	// initialize locker
	var locker handlertusd.Locker
	locker = memorylocker.New()
	if appConfig.TusRedisLockURI != "" {
		var err error
		locker, err = redislocker.New(appConfig.TusRedisLockURI, redislocker.WithLogger(logger))
		if err != nil {
			logger.Error("error configuring redis locker", "error", err)
			return nil, err
		}
	}

	// get and initialize tusd hook handlers
	hookHandler, err := GetHookHandler(appConfig)
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
	http.Handle(appConfig.TusdHandlerBasePath, http.StripPrefix(appConfig.TusdHandlerBasePath, handlerTusd))

	// initialize and route handler for DEX
	handlerDex := handlerdex.New(appConfig, psSender)
	http.Handle("/", handlerDex)

	// --------------------------------------------------------------
	// 	Prometheus metrics handler for /metrics
	// --------------------------------------------------------------
	hooks.SetupHookMetrics()
	http.Handle("/metrics", promhttp.Handler())
	setupMetrics()

	return http.DefaultServeMux, nil
}
