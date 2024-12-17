package handlertusd

import (
	"errors"
	"os"

	"golang.org/x/exp/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/slogerxexp"
	"github.com/prometheus/client_golang/prometheus"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/prometheuscollector"
) // .import

var logger *slog.Logger

func init() {
	logger = slogerxexp.With()
}

type Store interface {
	tusd.DataStore
	UseIn(*tusd.StoreComposer)
}

type Locker interface {
	tusd.Locker
	UseIn(*tusd.StoreComposer)
}

// New returns a configured TUSD handler as-is with official implementation
func New(store Store, locker Locker, hooksHandler hooks.HookHandler, basePath string) (*tusd.Handler, error) {
	if slogerxexp.DefaultLogger != nil {
		logger = slogerxexp.DefaultLogger
	}
	if store == nil {
		return nil, errors.New("no store provided")
	}
	if locker == nil {
		return nil, errors.New("no locker provided")
	}

	// tusd.Handler exposes metrics by cli flag and defaults true
	var handler *tusd.Handler

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	locker.UseIn(composer)

	// ------------------------------------------------------------------
	//  handler, set with respective local or cloud values
	// ------------------------------------------------------------------

	// Create a new HTTP handler for the tusd server by providing a configuration.
	// The StoreComposer property must be set to allow the handler to function.
	handler, err := hooks.NewHandlerWithHooks(&tusd.Config{
		BasePath:                   basePath,
		StoreComposer:              composer,
		NotifyCompleteUploads:      true,
		Logger:                     slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError})),
		RespectForwardedHeaders:    true,
		DisableDownload:            true,
		EnableExperimentalProtocol: true,
	}, hooksHandler, []hooks.HookType{hooks.HookPreCreate, hooks.HookPostCreate, hooks.HookPostReceive, hooks.HookPreFinish, hooks.HookPostFinish, hooks.HookPostTerminate}) // .handler
	if err != nil {
		logger.Error("error start tusd handler", "error", err)
		return nil, err
	} // .if

	if err := prometheus.Register(prometheuscollector.New(handler.Metrics)); err != nil && !errors.As(err, &prometheus.AlreadyRegisteredError{}) {
		return handler, err
	}
	logger.Info("started tusd handler")
	return handler, nil
} // .New
