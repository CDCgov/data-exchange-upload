package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/ui"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
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
	ctx, cancelFunc := context.WithCancel(context.Background())
	var mainWaitGroup sync.WaitGroup

	logger.Info("starting app")

	// Pub Sub
	event.MaxRetries = appConfig.EventMaxRetryCount
	// initialize event reporter
	err := cli.InitReporters(ctx, appConfig)
	defer reports.DefaultReporter.Close()

	event.InitFileReadyChannel()
	defer event.CloseFileReadyChannel()

	err = event.InitFileReadyPublisher(ctx, appConfig)
	defer event.FileReadyPublisher.Close()
	if err != nil {
		logger.Error("error creating file ready publisher", "error", err)
		os.Exit(appMainExitCode)
	}

	mainWaitGroup.Add(1)
	subscriber, err := cli.NewEventSubscriber[*event.FileReady](ctx, appConfig)
	if err != nil {
		logger.Error("error subscribing to file ready", "error", err)
		os.Exit(appMainExitCode)
	}
	defer subscriber.Close()
	go func() {
		cli.SubscribeToEvents(ctx, subscriber, postprocessing.ProcessFileReadyEvent)
		mainWaitGroup.Done()
	}()

	// start serving the app
	handler, err := cli.Serve(ctx, appConfig)
	if err != nil {
		logger.Error("error starting app, error initialize dex handler", "error", err)
		os.Exit(appMainExitCode)
	}

	logger.Info("http handlers ready")
	// ------------------------------------------------------------------
	// Start http custom server
	// ------------------------------------------------------------------
	httpServer := http.Server{

		Addr: ":" + appConfig.ServerPort,

		Handler: metrics.TrackHTTP(handler),
		// etc...

	} // .httpServer

	mainWaitGroup.Add(1)
	go func() {
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error starting app, error starting http server", "error", err, "port", appConfig.ServerPort)
			os.Exit(appMainExitCode)
		} // .if
		mainWaitGroup.Done()
	}() // .go

	logger.Info("started http server with tusd and dex handlers", "port", appConfig.ServerPort)

	mainWaitGroup.Add(1)
	go func() {
		defer mainWaitGroup.Done()
		if err := ui.Start(appConfig.UIPort, appConfig.TusUIFileEndpointUrl, appConfig.TusUIInfoEndpointUrl); err != nil {
			logger.Error("failed to start ui", "error", err)
			os.Exit(appMainExitCode)
		}
	}()

	logger.Info("Started ui server", "port", appConfig.UIPort)

	// ------------------------------------------------------------------
	// 	Block for Exit, server above is on goroutine
	// ------------------------------------------------------------------
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint
	cancelFunc()
	// ------------------------------------------------------------------
	// close other connections, if needed
	// ------------------------------------------------------------------
	httpShutdownCtx, httpShutdownCancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpShutdownCancelFunc()
	httpServer.Shutdown(httpShutdownCtx)
	ui.Close(httpShutdownCtx)

	mainWaitGroup.Wait()

	logger.Info("closing server by os signal", "port", appConfig.ServerPort)
} // .main
