package cli

import (
	"context"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sync"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/middleware"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	otrace "go.opentelemetry.io/otel/trace"
)

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

	// Register the trace exporter with a tracer provider, using a batch
	// span processor to aggregate spans before export
	bsp := trace.NewBatchSpanProcessor(traceExporter)
	var rngSeed int64
	binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithResource(res),
		trace.WithSpanProcessor(bsp),
		trace.WithIDGenerator(&UploadTraceIDGenerator{
			randSource: rand.New(rand.NewSource(rngSeed)),
		}),
	)
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (the default is no-op)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return traceExporter.Shutdown, nil
}

type UploadTraceIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
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
		u.randSource.Read(sid[:])
		if sid.IsValid() {
			break
		}
	}
	return sid
}

func (u *UploadTraceIDGenerator) newTraceID(ctx context.Context) otrace.TraceID {
	id := ctx.Value(middleware.UploadID)
	tid, ok := id.(otrace.TraceID)
	if ok {
		return tid
	}

	u.Lock()
	defer u.Unlock()
	tid = otrace.TraceID{}
	for {
		u.randSource.Read(tid[:])
		if tid.IsValid() {
			break
		}
	}

	return tid
}

func TracingProcessor[T event.Identifiable](next func(context.Context, T) error) func(context.Context, T) error {
	tracer := otel.Tracer("event-handling")
	return func(ctx context.Context, e T) error {
		c := context.WithValue(ctx, middleware.UploadID, otrace.TraceID(md5.Sum([]byte(e.GetUploadID()))))
		_, span := tracer.Start(c, fmt.Sprintf("Handling-%s", e.Identifier()))
		defer span.End()
		return next(c, e)
	}
}
