package handlertusd

import (
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/processingstatus"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/slogerxexp"
	"github.com/tus/tusd/v2/pkg/azurestore"
	"github.com/tus/tusd/v2/pkg/filelocker"
	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/memorylocker"
) // .import

// New returns a configured TUSD handler as-is with official implementation
func New(cliFlags cli.Flags, appConfig appconfig.AppConfig, psSender *processingstatus.PsSender) (*tusd.Handler, error) {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := slogerxexp.AppLogger(appConfig).With("pkg", pkgParts[len(pkgParts)-1])
	slogerxexp.SetDefaultLogger(logger)

	// tusd.Handler exposes metrics by cli flag and defaults true
	var handler *tusd.Handler
	var composer *tusd.StoreComposer

	logger.Info("loading environment for tusdhandler", "cliFlags.RunMode", cliFlags.RunMode)

	// ------------------------------------------------------------------
	// AZURE cloud
	// ------------------------------------------------------------------
	if cliFlags.RunMode == cli.RUN_MODE_AZURE || cliFlags.RunMode == cli.RUN_MODE_LOCAL_TO_AZURE {

		azConfig := &azurestore.AzConfig{
			AccountName:         appConfig.TusAzStorageConfig.AzStorageName,
			AccountKey:          appConfig.TusAzStorageConfig.AzStorageKey,
			ContainerName:       appConfig.TusAzStorageConfig.AzContainerName,
			ContainerAccessType: appConfig.TusAzStorageConfig.AzContainerAccessType,
			// BlobAccessTier:      Flags.AzBlobAccessTier,
			Endpoint: appConfig.TusAzStorageConfig.AzContainerEndpoint,
		} // .azConfig

		azService, err := azurestore.NewAzureService(azConfig)
		if err != nil {
			logger.Error("error create azure store service", "error", err)
			return nil, err
		} // azService

		store := azurestore.New(azService)
		// store.ObjectPrefix = Flags.AzObjectPrefix
		// store.Container = appConfig.AzContainerName

		// TODO: set for azure
		// TODO: set for azure, Upload Locks: https://tus.github.io/tusd/advanced-topics/locks/

		// A storage backend for tusd may consist of multiple different parts which
		// handle upload creation, locking, termination and so on. The composer is a
		// place where all those separated pieces are joined together. In this example
		// we only use the file store but you may plug in multiple.
		composer = tusd.NewStoreComposer()
		store.UseIn(composer)

		// used to prevent concurrent access to an upload: https://tus.github.io/tusd/advanced-topics/locks/
		// TODO: use azure cloud based locker
		// TODO: use azure cloud based locker
		locker := memorylocker.New()
		locker.UseIn(composer)

	} // .if cliFlags.RunMode == cliflags.RUN_MODE_AZURE

	// ------------------------------------------------------------------
	//  LOCAL is default if no flag is passed at cli
	// ------------------------------------------------------------------
	if cliFlags.RunMode == cli.RUN_MODE_LOCAL {

		// Create a new FileStore instance which is responsible for
		// storing the uploaded file on disk in the specified directory.
		// This path _must_ exist before tusd will store uploads in it.
		// If you want to save them on a different medium, for example
		// a remote FTP server, you can implement your own storage backend
		// by implementing the tusd.DataStore interface.
		store := filestore.FileStore{
			Path: appConfig.LocalFolderUploadsTus,
		} // .store

		// A storage backend for tusd may consist of multiple different parts which
		// handle upload creation, locking, termination and so on. The composer is a
		// place where all those separated pieces are joined together. In this example
		// we only use the file store but you may plug in multiple.
		composer = tusd.NewStoreComposer()
		store.UseIn(composer)

		// used to prevent concurrent access to an upload: https://tus.github.io/tusd/advanced-topics/locks/
		// ok for local dev to use disk based storage
		locker := filelocker.New(appConfig.LocalFolderUploadsTus)
		locker.UseIn(composer)

	} // .if cliFlags.RunMode == cliflags.RUN_MODE_LOCAL

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
	}, cli.GetHookHandler(), []hooks.HookType{hooks.HookPreCreate, hooks.HookPostFinish}) // .handler
	if err != nil {
		logger.Error("error start tusd handler", "error", err)
		return nil, err
	} // .if

	logger.Info("started tusd handler")
	return handler, nil
} // .New
