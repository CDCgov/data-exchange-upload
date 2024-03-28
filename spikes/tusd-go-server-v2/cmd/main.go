package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/hooks"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/processingstatus"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/serverdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/joho/godotenv"
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

	// ------------------------------------------------------------------
	// used to run the app locally, it uploads files locally
	// ------------------------------------------------------------------
	if cliFlags.RunMode == cli.RUN_MODE_LOCAL || cliFlags.RunMode == cli.RUN_MODE_LOCAL_TO_AZURE {
		err = godotenv.Load(cliFlags.AppLocalConfigPath)
		if err != nil {
			slog.Error("error loading local configuration", "runMode", cliFlags.RunMode, "error", err)
			os.Exit(appMainExitCode)
		} // .if
	} // .if

	// ------------------------------------------------------------------
	// used to run the app locally, it uploads files from local to azure
	// ------------------------------------------------------------------
	// load the additional azure configuration from local config yaml
	if cliFlags.RunMode == cli.RUN_MODE_LOCAL_TO_AZURE {
		err := godotenv.Load(cliFlags.AzLocalConfigPath)
		if err != nil {
			slog.Error("error loading local configuration", "runMode", cliFlags.RunMode, "error", err)
			os.Exit(appMainExitCode)
		} // .if
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
	} // .if

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
