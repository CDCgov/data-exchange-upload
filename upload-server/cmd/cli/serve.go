package cli

import (
	"context"
	"net/http"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/redislockerhealth"
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

	manifestMetrics := metrics.NewManifestMetrics(
		"upload_manifest_count",
		"The count of uploads by certain keys in the manifiest",
		appConfig.Metrics.LabelsFromManifest...)
	setupMetrics(manifestMetrics.Counter)

	// get and initialize tusd hook handlers
	hookHandler, err := GetHookHandler(ctx, appConfig)
	if err != nil {
		logger.Error("error configuring tusd handler: ", "error", err)
		return nil, err
	}
	hookHandler.Register(hooks.HookPostCreate, metrics.ActiveUploadIncHook)
	hookHandler.Register(hooks.HookPreFinish, manifestMetrics.Hook, metrics.ActiveUploadDecHook)

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
	http.Handle("/", appconfig.Handler())
	http.Handle("/health", health.Handler())

	// --------------------------------------------------------------
	// 	Prometheus metrics handler for /metrics
	// --------------------------------------------------------------
	http.Handle("/metrics", promhttp.Handler())

	http.Handle("/info/{UploadID}", uploadInfoHandler)
	http.Handle("/version", &VersionHandler{})
	http.Handle("/route/{UploadID}", &Router{})

	return http.DefaultServeMux, nil
}
