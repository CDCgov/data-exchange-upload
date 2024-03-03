package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/dexmetadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/serverdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
) // .import

const appMainExitCode = 1

func main() {

	buildInfo, _ := debug.ReadBuildInfo()

	// ------------------------------------------------------------------
	// parse and load config
	// ------------------------------------------------------------------
	appConfig, err := appconfig.ParseConfig()
	if err != nil {
		slog.Error("error starting app, error parsing app config", "error", err, "buildInfo.Main.Path", buildInfo.Main.Path)
		os.Exit(appMainExitCode)
	} // .if

	// ------------------------------------------------------------------
	// configure app custom logging
	// ------------------------------------------------------------------
	logger := sloger.AppLogger(appConfig)

	logger.Info("started app", "buildInfo.Main.Path", buildInfo.Main.Path)

	// TODO: context object, decide if custom slog is to be passed using the go context object
	// OR an internal logger package is to be used.

	// ------------------------------------------------------------------
	// parse and load cli flags
	// ------------------------------------------------------------------
	cliFlags, err := cliflags.ParseFlags()
	if err != nil {
		logger.Error("error starting app, error parsing cli flags", "error", err)
		os.Exit(appMainExitCode)
	} // .if

	// ------------------------------------------------------------------
	// load metadata v1 config into singleton to check and have available
	// ------------------------------------------------------------------
	_, err = dexmetadatav1.Load() // discard as not needed now in main
	if err != nil {
		logger.Error("error starting app, metadata v1 config not available", "error", err)
		os.Exit(appMainExitCode)
	} // .err

	// ------------------------------------------------------------------
	// create dex server, includes tusd as-is handler + dex handler
	// ------------------------------------------------------------------
	serverDex, err := serverdex.New(cliFlags, appConfig)
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
	httpServer.Shutdown(context.Background())

	logger.Info("closing server by os signal", "port", appConfig.ServerPort)
} // .main
