package main

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/server"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/flags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/config"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/tusdhandler"

) // .import


func main() {

	// TODO: structured logging, decide if slog is used and config at global level with default outputs
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	// TODO: context object, decide if custom slog is to be passed using the go context object

	// ------------------------------------------------------------------
	// parse and load cli flags
	// ------------------------------------------------------------------
	flags, err := flags.ParseFlags()
	if err != nil {
		logger.Error("error starting service, error parsing cli flags", "error", err)
		os.Exit(1)
	} // .if

	// ------------------------------------------------------------------
	// parse and load config
	// ------------------------------------------------------------------
	config, err := config.ParseConfig()
	if err != nil {
		logger.Error("error starting service, error parsing config", "error", err)
		os.Exit(1)
	} // .if

	// ------------------------------------------------------------------
	// create tusd handler
	// ------------------------------------------------------------------
	tusdHandler, err := tusdhandler.New(flags, config)
	if err != nil {
		logger.Error("error starting service and tusd handler", "error", err)
		os.Exit(1)
	} // .if

	// ------------------------------------------------------------------
	// create custom http server including tusd handler
	// ------------------------------------------------------------------
	httpServer, err := server.New(flags, config, tusdHandler)
	if err != nil {
		logger.Error("error starting service and http server", "error", err)
		os.Exit(1)
	} // .if

	// ------------------------------------------------------------------
	// Start http custom server, including tusd handler
	// ------------------------------------------------------------------
	logger.Info("starting http server, including tusd handler", "port", config.ServerPort)

	go func() {
		
		err := httpServer.Serve()
		if err != nil {
			logger.Error("error starting service, error starting http custom server", "error", err)
			os.Exit(1)
		} // .if

	}() // .go

	// ------------------------------------------------------------------
	// 			Block for Exit, everything above is on goroutines
	// ------------------------------------------------------------------
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	<-sigint

	// ------------------------------------------------------------------
	// close connections, TODO if needed
	// -----------------------------------------------------------------

	logger.Info("closing server by os signal", "port", config.ServerPort)

} // .main