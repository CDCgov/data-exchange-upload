package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/serverdex"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/slogerxexp"
	"github.com/joho/godotenv"
) // .import

const appMainExitCode = 1

var (
	appConfig appconfig.AppConfig
	logger    *slog.Logger
)

// NOTE: this large init file may be an antipattern.
// A main reason for it is to enable to cross cutting logging aspect.
// If another way is found to manage that this should be moved to main.
func init() {
	ctx := context.Background()

	buildInfo, _ := debug.ReadBuildInfo()
	logInfo := []any{"buildInfo.Main.Path", buildInfo.Main.Path}
	slog.With(logInfo...)
	// ------------------------------------------------------------------
	// parse and load cli flags
	// ------------------------------------------------------------------
	if !testing.Testing() {
		if err := cli.ParseFlags(); err != nil {
			slog.Error("error starting app, error parsing cli flags", "error", err)
			os.Exit(appMainExitCode)
		} // .if
	}

	if cli.Flags.AppConfigPath != "" {
		slog.Info("Loading environment from", "file", cli.Flags.AppConfigPath)
		if err := godotenv.Load(cli.Flags.AppConfigPath); err != nil {
			slog.Error("error loading local configuration", "error", err)
			os.Exit(appMainExitCode)
		} // .if
	}

	// ------------------------------------------------------------------
	// parse and load config from os exported
	// ------------------------------------------------------------------
	var err error
	appConfig, err = appconfig.ParseConfig(ctx)
	if err != nil {
		slog.Error("error starting app, error parsing app config", "error", err)
		os.Exit(appMainExitCode)
	} // .if

	// ------------------------------------------------------------------
	// configure app custom logging
	// ------------------------------------------------------------------
	logInfo = append(logInfo, "pkg", "main")
	logger = cli.AppLogger(appConfig).With(logInfo...)
	sloger.SetDefaultLogger(logger)

	explogger := cli.ExpAppLogger(appConfig).With(logInfo...)
	slogerxexp.SetDefaultLogger(explogger)

}

func main() {
	ctx := context.Background()

	logger.Info("starting app")

	// start serving the app
	_, err := cli.Serve(appConfig)
	if err != nil {
		logger.Error("error starting app, error initialize dex handler", "error", err)
		os.Exit(appMainExitCode)
	}

	logger.Info("http handlers ready")
	// ------------------------------------------------------------------
	// create dex server, includes dex handler
	// ------------------------------------------------------------------
	serverDex, err := serverdex.New(appConfig)
	if err != nil {
		logger.Error("error starting app, error initialize dex server", "error", err)
		os.Exit(appMainExitCode)
	} // .if

	logger.Info("http server ready")

	// ------------------------------------------------------------------
	// Start http custom server
	// ------------------------------------------------------------------
	httpServer := serverDex.HttpServer()

	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
