package cli

import (
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/processingstatus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/memorylocker"
)

func Serve(appConfig appconfig.AppConfig) (http.Handler, error) {

	// ------------------------------------------------------------------
	// load metadata v1 config into singleton to check and have available
	// ------------------------------------------------------------------
	_, err := metadatav1.LoadOnce(appConfig)
	if err != nil {
		logger.Error("error starting app, metadata v1 config not available", "error", err)
		return nil, err
	} // .err

	psSender, err := processingstatus.New(appConfig)
	if err != nil {
		logger.Error("error processing status not available", "error", err)
	} // .err
	if psSender != nil {
		health.Register(psSender)
	}

	store, storeHealthCheck, err := CreateDataStore(appConfig)
	if err != nil {
		logger.Error("error starting app, error configuring storage", "error", err)
		return nil, err
	}
	health.Register(storeHealthCheck)

	locker := memorylocker.New()

	handlerTusd, err := handlertusd.New(store, locker, GetHookHandler(), appConfig.TusdHandlerBasePath)
	if err != nil {
		logger.Error("error starting tusd handler: ", err)
		return nil, err
	} // .handlerTusd

	// --------------------------------------------------------------
	// 	TUSD handler
	// --------------------------------------------------------------
	// Route for TUSD to start listening on and accept http request
	http.Handle(appConfig.TusdHandlerBasePath, http.StripPrefix(appConfig.TusdHandlerBasePath, handlerTusd))

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
