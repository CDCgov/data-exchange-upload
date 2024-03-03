package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/dexmetadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/serverdex"
) // .import

func main() {

	// TODO: structured logging, decide if slog is used and config at global level with default outputs
	loggerHandler := slog.NewJSONHandler(os.Stdout, nil)

	// buildInfo, _ := debug.ReadBuildInfo()

	parentLogger := slog.New(loggerHandler)

	logger := parentLogger.With(
		slog.Group("app_info",
			slog.String("System", "OCIO DEX"), // TODO: can come from config
			slog.String("Product", "Upload API"),
			slog.String("App", "tusd-go-server"),
			slog.Int("pid", os.Getpid()),
		),
	)

	logger.Info("starting application...")

	// TODO: context object, decide if custom slog is to be passed using the go context object
	// OR an internal logger package is to be used.

	// ------------------------------------------------------------------
	// parse and load cli flags
	// ------------------------------------------------------------------
	cliFlags, err := cliflags.ParseFlags()
	if err != nil {
		logger.Error("error starting service, error parsing cli flags", "error", err)
		os.Exit(1)
	} // .if

	// ------------------------------------------------------------------
	// parse and load config
	// ------------------------------------------------------------------
	appConfig, err := appconfig.ParseConfig()
	if err != nil {
		logger.Error("error starting service, error parsing config", "error", err)
		os.Exit(1)
	} // .if

	// ------------------------------------------------------------------
	// load metadata v1 config into singleton to check and have available
	// ------------------------------------------------------------------
	_, err = dexmetadatav1.Load() // discard as not needed now in main
	if err != nil {
		logger.Error("error metadata v1 config not available", "error", err)
		os.Exit(1)
	} // .err

	// ------------------------------------------------------------------
	// create custom http server, includes tusd as-is handler + dex handler
	// ------------------------------------------------------------------
	serverDex, err := serverdex.New(cliFlags, appConfig)
	if err != nil {
		logger.Error("error starting service and http server", "error", err)
		os.Exit(1)
	} // .if

	// ------------------------------------------------------------------
	// Start http custom server, including tusd handler
	// ------------------------------------------------------------------
	logger.Info("starting http server, including tusd handler", "port", appConfig.ServerPort)

	go func() {

		httpServer := serverDex.HttpServer()
		err := httpServer.ListenAndServe()
		if err != nil {
			logger.Error("error starting service, error starting http custom server", "error", err)
			os.Exit(1)
		} // .if

	}() // .go

	// ------------------------------------------------------------------
	// 	Block for Exit, server above is on goroutine
	// ------------------------------------------------------------------
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	<-sigint

	// ------------------------------------------------------------------
	// close connections, TODO if needed
	// -----------------------------------------------------------------

	logger.Info("closing server by os signal", "port", appConfig.ServerPort)
} // .main
