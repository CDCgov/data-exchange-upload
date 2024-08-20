package cli

import (
	"context"
	"net/http"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/redislocker"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/memorylocker"
)

func Serve(ctx context.Context, appConfig appconfig.AppConfig) (http.Handler, error) {
	if sloger.DefaultLogger != nil {
		logger = sloger.DefaultLogger
	}

	// create and register data store
	store, storeHealthCheck, err := GetDataStore(ctx, appConfig)
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
			return nil, err
		}
		health.Register(locker.(health.Checkable))
	}

	err = postprocessing.RegisterAllTargets(ctx, appConfig)
	if err != nil {
		return nil, err
	}

	// Must be called before hook handler
	err = metadata.InitConfigCache(ctx, appConfig)
	if err != nil {
		return nil, err
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
	http.Handle("/route/{UploadID}", &Router{})

	return http.DefaultServeMux, nil
}
