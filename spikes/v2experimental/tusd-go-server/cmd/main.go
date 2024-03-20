package main

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/serverdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
) // .import

const appMainExitCode = 1

func main() {

	ctx := context.Background()

	buildInfo, _ := debug.ReadBuildInfo()

	// ------------------------------------------------------------------
	// parse and load cli flags
	// ------------------------------------------------------------------
	cliFlags, err := cliflags.ParseFlags()
	if err != nil {
		slog.Error("error starting app, error parsing cli flags", "error", err, "buildInfo.Main.Path", buildInfo.Main.Path)
		os.Exit(appMainExitCode)
	} // .if

	// ------------------------------------------------------------------
	// used to run the app locally, it uploads files locally
	// ------------------------------------------------------------------
	if cliFlags.Environment == cliflags.ENV_LOCAL || cliFlags.Environment == cliflags.ENV_LOCAL_TO_AZURE {
		err = godotenv.Load(cliFlags.AppLocalConfigPath)
		if err != nil {
			slog.Error("error loading local configuration", "environment", cliFlags.Environment, "error", err)
			os.Exit(appMainExitCode)
		} // .if
	} // .if

	// ------------------------------------------------------------------
	// used to run the app locally, it uploads files from local to azure
	// ------------------------------------------------------------------
	if cliFlags.Environment == cliflags.ENV_LOCAL_TO_AZURE {
		err := godotenv.Load(cliFlags.AzLocalConfigPath) 
		if err != nil {
			slog.Error("error loading local configuration", "environment", cliFlags.Environment, "error", err)
			os.Exit(appMainExitCode)
		} // .if
	} // .if

	// ------------------------------------------------------------------
	// parse and load config
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

	// ------------------------------------------------------------------
	// create dex server, includes tusd as-is handler + dex handler
	// ------------------------------------------------------------------
	serverDex, err := serverdex.New(cliFlags, appConfig, metaV1)
	if err != nil {
		logger.Error("error starting app, error initialize dex server", "error", err)
		os.Exit(appMainExitCode)
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
