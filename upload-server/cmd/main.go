package main

import (
	"context"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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
	"github.com/google/uuid"

	"github.com/cdcgov/data-exchange-upload/upload-server/cmd/cli"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	otrace "go.opentelemetry.io/otel/trace"
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

// Initializes an OTLP exporter, and configures the corresponding trace provider.
func initTracerProvider(ctx context.Context) (func(context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// The service name used to display traces in backends
			semconv.ServiceNameKey.String("upload-server"),
		),
	)
	if err != nil {
		return nil, err
	}
	// Set up a trace exporter
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := trace.NewBatchSpanProcessor(traceExporter)
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
		trace.WithIDGenerator(&UploadTraceIDGenerator{
			randSource: rand.New(rand.NewSource(rngSeed)),
		}),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

type UploadTraceIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

func (u *UploadTraceIDGenerator) newTraceID(ctx context.Context) otrace.TraceID {

	id := ctx.Value(cli.UploadID)
	tid, ok := id.(otrace.TraceID)
	if ok {
		return tid
	}

	u.Lock()
	defer u.Unlock()
	tid = otrace.TraceID{}
	for {
		_, _ = u.randSource.Read(tid[:])
		if tid.IsValid() {
			break
		}
	}
	return tid
}

func (u *UploadTraceIDGenerator) NewIDs(ctx context.Context) (otrace.TraceID, otrace.SpanID) {
	tid := u.newTraceID(ctx)
	sid := u.NewSpanID(ctx, tid)
	return tid, sid
}

func (u *UploadTraceIDGenerator) NewSpanID(ctx context.Context, traceID otrace.TraceID) otrace.SpanID {
	u.Lock()
	defer u.Unlock()
	sid := otrace.SpanID{}
	for {
		_, _ = u.randSource.Read(sid[:])
		if sid.IsValid() {
			break
		}
	}
	return sid
}

func main() {
	var mainWaitGroup sync.WaitGroup
	ctx, cancelFunc := context.WithCancel(context.Background())
	tpShutdown, err := initTracerProvider(ctx)
	if err != nil {
		slog.Error("error starting app, error starting tracing", "error", err)
		os.Exit(appMainExitCode)
	}

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

	if err := event.InitFileReadyPublisher(ctx, appConfig); err != nil {
		slog.Error("error creating file ready publisher", "error", err)
		os.Exit(appMainExitCode)
	}
	defer event.FileReadyPublisher.Close()

	mainWaitGroup.Add(1)
	subscriber, err := cli.NewEventSubscriber[*event.FileReady](ctx, appConfig)
	if err != nil {
		slog.Error("error subscribing to file ready", "error", err)
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
		slog.Error("error starting app, error initialize dex handler", "error", err)
		os.Exit(appMainExitCode)
	}

	slog.Info("http handlers ready")
	// ------------------------------------------------------------------
	// Start http custom server
	// ------------------------------------------------------------------

	uploadIDContext := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			suffix := filepath.Base(path)
			// TODO trim + size to allow for s3 case
			if err := uuid.Validate(suffix); err == nil {
				ctx := r.Context()
				id := otrace.TraceID(md5.Sum([]byte(suffix)))
				slog.Info("tracing upload", "upload_id", suffix, "trace_id", id.String())
				r = r.WithContext(context.WithValue(ctx, cli.UploadID, id))
			}
			next.ServeHTTP(rw, r)
		})
	}
	//TODO move this down a layer to trace metrics vs tus etc.
	tracingMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			otelhttp.NewMiddleware(fmt.Sprintf("%s %s", r.Method, r.URL.Path))(next).ServeHTTP(rw, r)
		})
	}
	httpServer := http.Server{

		Addr: ":" + appConfig.ServerPort,

		Handler: uploadIDContext(tracingMiddleware(metrics.TrackHTTP(handler))),
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
			if err := ui.Start(appConfig.UIPort, appConfig.CsrfToken, appConfig.ServerFileEndpointUrl, appConfig.ServerInfoEndpointUrl); err != nil {
				slog.Error("failed to start ui", "error", err)
				os.Exit(appMainExitCode)
			}
		}()

		slog.Info("Started ui server", "port", appConfig.UIPort)
	}

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
	tpShutdown(httpShutdownCtx)

	mainWaitGroup.Wait()

	slog.Info("closing server by os signal", "port", appConfig.ServerPort)
} // .main
