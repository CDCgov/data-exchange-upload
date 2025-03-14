package cli

import (
	"context"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	otrace "go.opentelemetry.io/otel/trace"
)

type key int

var UploadID key

// Initializes an OTLP exporter, and configures the corresponding trace provider.
func InitTracerProvider(ctx context.Context) (func(context.Context) error, error) {
	res, err := resource.New(ctx)
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

	id := ctx.Value(UploadID)
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

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		otelhttp.NewMiddleware(fmt.Sprintf("%s %s", r.Method, r.URL.Path))(next).ServeHTTP(rw, r)
	})
}

func AddUploadIDContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		suffix := filepath.Base(path)
		// TODO trim + size to allow for s3 case
		if err := uuid.Validate(suffix); err == nil {
			ctx := r.Context()
			id := otrace.TraceID(md5.Sum([]byte(suffix)))
			slog.Info("tracing upload", "upload_id", suffix, "trace_id", id.String())
			r = r.WithContext(context.WithValue(ctx, UploadID, id))
		}
		next.ServeHTTP(rw, r)
	})
}

func TracingProcessor[T event.Identifiable](next func(context.Context, T) error) func(context.Context, T) error {
	tracer := otel.Tracer("event-handling")
	return func(ctx context.Context, e T) error {
		c := context.WithValue(ctx, UploadID, otrace.TraceID(md5.Sum([]byte(e.GetUploadID()))))
		_, span := tracer.Start(c, fmt.Sprintf("Handling-%s", e.Identifier()))
		defer span.End()
		return next(c, e)
	}
}
