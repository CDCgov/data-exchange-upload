package cli

import (
	"context"
	"net/http"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/middleware"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/redislocker"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/memorylocker"
)

func Serve(ctx context.Context, appConfig appconfig.AppConfig, authMiddleware *middleware.AuthMiddleware) (http.Handler, error) {
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

	uploadInfoHandler, err := GetUploadInfoHandler(ctx, &appConfig)
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
			logger.Error("failed to initialize Redis Locker", "error", err)
			return nil, err
		}
		health.Register(locker.(health.Checkable))
	}

	manifestMetrics := metrics.NewManifestMetrics(
		"upload_manifest_count",
		"The count of uploads by certain keys in the manifiest",
		appConfig.Metrics.LabelsFromManifest...)
	setupMetrics(ctx, manifestMetrics.Counter)

	// Must be called before hook handler
	err = InitConfigCache(ctx, appConfig)
	if err != nil {
		return nil, err
	}

	err = RegisterAllSourcesAndDestinations(ctx, appConfig)
	if err != nil {
		return nil, err
	}

	// get and initialize tusd hook handlers
	hookHandler, err := GetHookHandler(appConfig)
	if err != nil {
		logger.Error("error configuring tusd handler: ", "error", err)
		return nil, err
	}
	hookHandler.Register(hooks.HookPostCreate, metrics.ActiveUploadIncHook)
	hookHandler.Register(hooks.HookPostFinish, manifestMetrics.Hook, metrics.ActiveUploadDecHook, metrics.UploadSpeedsHook)

	// initialize tusd handler
	handlerTusd, err := handlertusd.New(store, locker, hookHandler, appConfig.TusdHandlerBasePath)
	if err != nil {
		logger.Error("error starting tusd handler: ", "error", err)
		return nil, err
	}

	mux := &http.ServeMux{}
	// --------------------------------------------------------------
	// 	TUSD handler
	// --------------------------------------------------------------

	// Route for TUSD to start listening on and accept http request
	logger.Info("hosting tus handler", "path", appConfig.TusdHandlerBasePath)
	pathWithoutSlash := strings.TrimSuffix(appConfig.TusdHandlerBasePath, "/")
	pathWithSlash := pathWithoutSlash + "/"
	mux.Handle(pathWithoutSlash, authMiddleware.VerifyOAuthTokenMiddleware(http.StripPrefix(pathWithoutSlash, handlerTusd)))
	mux.Handle(pathWithSlash, authMiddleware.VerifyOAuthTokenMiddleware(http.StripPrefix(pathWithSlash, handlerTusd)))

	// initialize and route handler for DEX
	mux.Handle("/health", health.Handler())

	// --------------------------------------------------------------
	// 	Prometheus metrics handler for /metrics
	// --------------------------------------------------------------
	mux.Handle("/metrics", promhttp.Handler())

	mux.Handle("/info/{UploadID}", authMiddleware.VerifyOAuthTokenMiddleware(uploadInfoHandler))
	mux.Handle("/version", &VersionHandler{})
	mux.Handle("/route/{UploadID}", &Router{})

	mux.Handle("/{$}", appconfig.Handler())

	return mux, nil
}
