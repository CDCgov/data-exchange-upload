package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/hooks"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/processingstatus"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/serverdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/joho/godotenv"
	"github.com/tus/tusd/v2/pkg/azurestore"
	"github.com/tus/tusd/v2/pkg/filelocker"
	"github.com/tus/tusd/v2/pkg/filestore"
) // .import

const appMainExitCode = 1

func main() {

	ctx := context.Background()

	buildInfo, _ := debug.ReadBuildInfo()

	// ------------------------------------------------------------------
	// parse and load cli flags
	// ------------------------------------------------------------------
	cliFlags, err := cli.ParseFlags()
	if err != nil {
		slog.Error("error starting app, error parsing cli flags", "error", err, "buildInfo.Main.Path", buildInfo.Main.Path)
		os.Exit(appMainExitCode)
	} // .if

	if err := godotenv.Load(cliFlags.AppConfigPath); err != nil {
		slog.Error("error loading local configuration", "runMode", cliFlags.RunMode, "error", err)
		os.Exit(appMainExitCode)
	} // .if

	// ------------------------------------------------------------------
	// parse and load config from os exported
	// ------------------------------------------------------------------
	appConfig, err := appconfig.ParseConfig(ctx)
	if err != nil {
		slog.Error("error starting app, error parsing app config", "error", err, "buildInfo.Main.Path", buildInfo.Main.Path)
		os.Exit(appMainExitCode)
	} // .if

	// ------------------------------------------------------------------
	// configure app custom logging
	// ------------------------------------------------------------------
	logger := sloger.AppLogger(appConfig).With("pkg", "main")

	logger.Info("started app", "buildInfo.Main.Path", buildInfo.Main.Path)

	// logger.Debug("loaded app config", "appConfig", appConfig)

	// ------------------------------------------------------------------
	// load metadata v1 config into singleton to check and have available
	// ------------------------------------------------------------------
	metaV1, err := metadatav1.LoadOnce(appConfig)
	if err != nil {
		logger.Error("error starting app, metadata v1 config not available", "error", err)
		os.Exit(appMainExitCode)
	} // .err

	psSender, err := processingstatus.New(appConfig)
	if err != nil {
		logger.Error("error processing status not available", "error", err)
	} // .err

	// ------------------------------------------------------------------
	// create dex server, includes tusd as-is handler + dex handler
	// ------------------------------------------------------------------
	serverDex, err := serverdex.New(cliFlags, appConfig, metaV1, psSender)
	if err != nil {
		logger.Error("error starting app, error initialize dex server", "error", err)
		os.Exit(appMainExitCode)
	} // .if

	var store handlertusd.Store
	var locker handlertusd.Locker
	// ------------------------------------------------------------------
	// Load Az dependencies, needed for the DEX handler paths
	// ------------------------------------------------------------------
	if cliFlags.RunMode == cli.RUN_MODE_LOCAL_TO_AZURE || cliFlags.RunMode == cli.RUN_MODE_AZURE {
		// load on server azure service dependencies

		// TODO: create the extra container that tus blob client needs it: one for raw uploads + one for dex uploads (files + manifest)
		serverDex.HandlerDex.TusAzBlobClient, err = storeaz.NewTusAzBlobClient(appConfig)
		if err != nil {
			logger.Error("error receive az tus blob client", "error", err)
		} // .if

		serverDex.HandlerDex.RouterAzBlobClient, err = storeaz.NewRouterAzBlobClient(appConfig)
		if err != nil {
			logger.Error("error receive az router blob client", "error", err)
		} // .if

		serverDex.HandlerDex.EdavAzBlobClient, err = storeaz.NewEdavAzBlobClient(appConfig)
		if err != nil {
			logger.Error("error receive az edav blob client", "error", err)
		} // .if

		azHook := &hooks.AzureUploadCompleteHandler{
			TusAzBlobClient:    serverDex.HandlerDex.TusAzBlobClient,
			RouterAzBlobClient: serverDex.HandlerDex.RouterAzBlobClient,
			EdavAzBlobClient:   serverDex.HandlerDex.EdavAzBlobClient,
		}
		cli.PostProcessHook = azHook.AzurePostProcess

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
			os.Exit(appMainExitCode)
		} // azService

		store = azurestore.New(azService)
		// store.ObjectPrefix = Flags.AzObjectPrefix
		// store.Container = appConfig.AzContainerName

		// TODO: set for azure
		// TODO: set for azure, Upload Locks: https://tus.github.io/tusd/advanced-topics/locks/
	} else { // .if
		// Create a new FileStore instance which is responsible for
		// storing the uploaded file on disk in the specified directory.
		// This path _must_ exist before tusd will store uploads in it.
		// If you want to save them on a different medium, for example
		// a remote FTP server, you can implement your own storage backend
		// by implementing the tusd.DataStore interface.
		store = filestore.FileStore{
			Path: appConfig.LocalFolderUploadsTus,
		} // .store

		// used to prevent concurrent access to an upload: https://tus.github.io/tusd/advanced-topics/locks/
		// ok for local dev to use disk based storage
		locker = filelocker.New(appConfig.LocalFolderUploadsTus)
	}

	handlerTusd, err := handlertusd.New(store, locker, cli.GetHookHandler(), appConfig)
	if err != nil {
		logger.Error("error starting tusd handler: ", err)
		os.Exit(appMainExitCode)
	} // .handlerTusd
	// --------------------------------------------------------------
	// 	TUSD handler
	// --------------------------------------------------------------
	// Route for TUSD to start listening on and accept http request
	http.Handle(appConfig.TusdHandlerBasePath, http.StripPrefix(appConfig.TusdHandlerBasePath, handlerTusd))

	// ------------------------------------------------------------------
	// Start http custom server
	// ------------------------------------------------------------------
	httpServer := serverDex.HttpServer()

	go func() {

		err := httpServer.ListenAndServe()
		if err != nil {
			logger.Error("error starting app, error starting http server", "error", err, "port", appConfig.ServerPort)
			os.Exit(appMainExitCode)
		} // .if
	}() // .go
	logger.Info("started http server with tusd and dex handlers", "port", appConfig.ServerPort)

	// ------------------------------------------------------------------
	// 	Block for Exit, server above is on goroutine
	// ------------------------------------------------------------------
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	<-sigint

	// ------------------------------------------------------------------
	// close other connections, if needed
	// ------------------------------------------------------------------
	httpServer.Shutdown(ctx)

	logger.Info("closing server by os signal", "port", appConfig.ServerPort)
} // .main
