package handlertusd

import (
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/slogerxexp"
	"github.com/prometheus/client_golang/prometheus"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/prometheuscollector"
) // .import

type Store interface {
	tusd.DataStore
	UseIn(*tusd.StoreComposer)
}

type Locker interface {
	tusd.Locker
	UseIn(*tusd.StoreComposer)
}

// New returns a configured TUSD handler as-is with official implementation
func New(store Store, locker Locker, hooksHandler hooks.HookHandler, appConfig appconfig.AppConfig) (*tusd.Handler, error) {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := slogerxexp.AppLogger(appConfig).With("pkg", pkgParts[len(pkgParts)-1])
	slogerxexp.SetDefaultLogger(logger)

	// tusd.Handler exposes metrics by cli flag and defaults true
	var handler *tusd.Handler
	var composer *tusd.StoreComposer

	composer = tusd.NewStoreComposer()
	store.UseIn(composer)
	locker.UseIn(composer)

	// ------------------------------------------------------------------
	//  handler, set with respective local or cloud values
	// ------------------------------------------------------------------

	// Create a new HTTP handler for the tusd server by providing a configuration.
	// The StoreComposer property must be set to allow the handler to function.
	handler, err := hooks.NewHandlerWithHooks(&tusd.Config{
		BasePath:              appConfig.TusdHandlerBasePath,
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
		// NotifyTerminatedUploads: true,
		// NotifyUploadProgress:    true,
		// NotifyCreatedUploads:    true,
		// PreUploadCreateCallback:

		// TODO: the tusd logger type is "golang.org/x/exp/slog" vs. app logger "log/slog"
		// TODO: switch to the log/slog when tusd is on that
		Logger: logger,
	}, hooksHandler, []hooks.HookType{hooks.HookPreCreate, hooks.HookPostFinish}) // .handler
	if err != nil {
		logger.Error("error start tusd handler", "error", err)
		return nil, err
	} // .if

	prometheus.MustRegister(prometheuscollector.New(handler.Metrics))
	logger.Info("started tusd handler")
	return handler, nil
} // .New
