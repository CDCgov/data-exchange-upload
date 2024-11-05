package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/ui"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/slogerxexp"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/joho/godotenv"
) // .import

const appMainExitCode = 1

func init() {
	// ------------------------------------------------------------------
	// parse and load cli flags
	// ------------------------------------------------------------------
	if !testing.Testing() {
		if err := cli.ParseFlags(); err != nil {
			slog.Error("error starting app, error parsing cli flags", "error", err)
			os.Exit(appMainExitCode)
		}
	}

	if cli.Flags.AppConfigPath != "" {
		slog.Info("Loading environment from", "file", cli.Flags.AppConfigPath)
		if err := godotenv.Load(cli.Flags.AppConfigPath); err != nil {
			slog.Error("error loading local configuration", "error", err)
			os.Exit(appMainExitCode)
		} // .if
	}

}

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	var mainWaitGroup sync.WaitGroup

	appConfig, err := appconfig.ParseConfig(ctx)
	if err != nil {
		slog.Error("error starting app, error parsing app config", "error", err)
		os.Exit(appMainExitCode)
	} // .if

	sloger.SetDefaultLogger(cli.AppLogger(appConfig))
	slog.SetDefault(sloger.DefaultLogger)
	slogerxexp.SetDefaultLogger(cli.ExpAppLogger(appConfig))
	slog.Info("starting app")

	// Pub Sub
	event.MaxRetries = appConfig.EventMaxRetryCount
	// initialize event reporter
	if err := cli.InitReporters(ctx, appConfig); err != nil {
		slog.Error("error creating reporters", "error", err)
		os.Exit(appMainExitCode)
	}
	defer reports.CloseAll()

	event.InitFileReadyChannel()
	defer event.CloseFileReadyChannel()

	if err := cli.InitFileReadyPublisher(ctx, appConfig); err != nil {
		slog.Error("error creating file ready publisher", "error", err)
		os.Exit(appMainExitCode)
	}
	defer event.FileReadyPublisher.Close()

	mainWaitGroup.Add(appConfig.ListenerWorkers)
	for range appConfig.ListenerWorkers {
		subscriber, err := cli.NewEventSubscriber[*event.FileReady](ctx, appConfig)
		if err != nil {
			slog.Error("error subscribing to file ready", "error", err)
			os.Exit(appMainExitCode)
		}
		if sc, ok := subscriber.(interface {
			Close() error
		}); ok {
			defer sc.Close()
		}
		go func() {
			defer mainWaitGroup.Done()
			if err := subscriber.Listen(ctx, postprocessing.ProcessFileReadyEvent); err != nil {
				cancelFunc()
				slog.Error("Listener failed", "error", err)
			}
		}()
	}

	// start serving the app
	handler, err := cli.Serve(ctx, appConfig)
	if err != nil {
		slog.Error("error starting app, error initialize dex handler", "error", err)
		os.Exit(appMainExitCode)
	}

	slog.Info("http handlers ready")
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
			slog.Error("error starting app, error starting http server", "error", err, "port", appConfig.ServerPort)
			os.Exit(appMainExitCode)
		} // .if
		mainWaitGroup.Done()
	}() // .go

	slog.Info("started http server with tusd and dex handlers", "port", appConfig.ServerPort)

	if appConfig.UIPort != "" {
		mainWaitGroup.Add(1)
		go func() {
			defer mainWaitGroup.Done()
			if err := ui.Start(appConfig.UIPort, appConfig.CsrfToken, appConfig.ExternalServerFileEndpointUrl, appConfig.InternalServerInfoEndpointUrl, appConfig.InternalServerFileEndpointUrl); err != nil {
				slog.Error("failed to start ui", "error", err)
				os.Exit(appMainExitCode)
			}
		}()

		slog.Info("Started ui server", "port", appConfig.UIPort)
	}

	// ------------------------------------------------------------------
	// 	Block for Exit, server above is on goroutine
	// ------------------------------------------------------------------
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		cancelFunc()
	}()
	<-ctx.Done()
	// ------------------------------------------------------------------
	// close other connections, if needed
	// ------------------------------------------------------------------
	httpShutdownCtx, httpShutdownCancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer httpShutdownCancelFunc()
	httpServer.Shutdown(httpShutdownCtx)
	ui.Close(httpShutdownCtx)

	mainWaitGroup.Wait()

	slog.Info("closing server by os signal", "port", appConfig.ServerPort)
} // .main
