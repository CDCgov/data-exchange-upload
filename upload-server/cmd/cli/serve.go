package cli

import (
	"context"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"net/http"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
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
)

func Serve(appConfig appconfig.AppConfig) (http.Handler, error) {
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

	// Get one or more health checks for delivery stores.
	// Maybe register deliverers at this level, and register health checks for them.
	// Don't need to call postprocessing.RegisterTarget in hooks.go.  Can do it here if we want.
	// Something like...
	// dexDeliverer = postProcessing.GetDeliverer("dex", appConfig)
	// health.Register(dexDeliverer)
	// postProcessing.RegisterDeliverer(dexDeliverer)
	ctx := context.TODO()
	// TODO can this just be of type deliverer interface?
	dexFileDeliverer, err := postprocessing.NewFileDeliverer(ctx, "dex")
	if err != nil {
		return nil, err
	}
	if appConfig.AzureConnection != nil {
		dexAzureDeliverer, err := postprocessing.NewAzureDeliverer(ctx, "dex", &appConfig)
		if err != nil {
			return nil, err
		}
		postprocessing.RegisterTarget("dex", dexAzureDeliverer)
		health.Register(dexAzureDeliverer)
	} else {
		fmt.Printf("***Registering file deliverer for dex")
		postprocessing.RegisterTarget("dex", dexFileDeliverer)
		health.Register(dexFileDeliverer)
	}

	edavFileDeliverer, err := postprocessing.NewFileDeliverer(ctx, "edav")
	if err != nil {
		return nil, err
	}
	if appConfig.EdavConnection != nil {
		edavAzureDeliverer, err := postprocessing.NewAzureDeliverer(ctx, "edav", &appConfig)
		if err != nil {
			return nil, err
		}
		postprocessing.RegisterTarget("edav", edavAzureDeliverer)
		health.Register(edavAzureDeliverer)
	} else {
		postprocessing.RegisterTarget("edav", edavFileDeliverer)
		health.Register(edavFileDeliverer)
	}

	routingFileDeliverer, err := postprocessing.NewFileDeliverer(ctx, "routing")
	if err != nil {
		return nil, err
	}
	if appConfig.RoutingConnection != nil {
		routingAzureDeliverer, err := postprocessing.NewAzureDeliverer(ctx, "routing", &appConfig)
		if err != nil {
			return nil, err
		}
		postprocessing.RegisterTarget("routing", routingAzureDeliverer)
		health.Register(routingAzureDeliverer)
	} else {
		postprocessing.RegisterTarget("routing", routingFileDeliverer)
		health.Register(routingFileDeliverer)
	}

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
			logger.Error("error configuring redis locker", "error", err)
			return nil, err
		}
		// redislocker health check
		redisLockerHealth, err := redislockerhealth.New(appConfig.TusRedisLockURI)
		if err != nil {
			logger.Error("error configuring redis locker health check: ", "error", err)
			return nil, err
		}

		health.Register(redisLockerHealth)
	}

	// initialize event reporter
	err = InitReporters(appConfig)

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

	return http.DefaultServeMux, nil
}
