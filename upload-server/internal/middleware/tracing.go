package middleware

import (
	"context"
	"crypto/md5"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

type key int

var UploadID key

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
			id := trace.TraceID(md5.Sum([]byte(suffix)))
			slog.Info("tracing upload", "upload_id", suffix, "trace_id", id.String())
			r = r.WithContext(context.WithValue(ctx, UploadID, id))
		}
		next.ServeHTTP(rw, r)
	})
}
