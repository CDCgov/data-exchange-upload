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
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/processingstatus"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/serverdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/joho/godotenv"
	"github.com/tus/tusd/v2/pkg/memorylocker"
) // .import

const appMainExitCode = 1

func main() {

	ctx := context.Background()

	buildInfo, _ := debug.ReadBuildInfo()

	// ------------------------------------------------------------------
	// parse and load cli flags
	// ------------------------------------------------------------------
	err := cli.ParseFlags()
	if err != nil {
		slog.Error("error starting app, error parsing cli flags", "error", err, "buildInfo.Main.Path", buildInfo.Main.Path)
		os.Exit(appMainExitCode)
	} // .if

	if err := godotenv.Load(cli.Flags.AppConfigPath); err != nil {
		slog.Error("error loading local configuration", "runMode", cli.Flags.RunMode, "error", err)
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

	store, err := cli.CreateDataStore(appConfig)
	if err != nil {
		logger.Error("error starting app, error configuring storage", "error", err)
		os.Exit(appMainExitCode)
	}

	locker := memorylocker.New()

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
	// create dex server, includes dex handler
	// ------------------------------------------------------------------
	serverDex, err := serverdex.New(appConfig, metaV1, psSender)
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
