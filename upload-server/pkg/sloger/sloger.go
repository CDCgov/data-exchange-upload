package sloger

import (
	"context"
	"log/slog"
)

type ContextKey string

var LoggerKey ContextKey = "logger"

var (
	DefaultLogger = slog.Default()
)

func SetDefaultLogger(l *slog.Logger) {
	DefaultLogger = l
}

func With(args ...any) *slog.Logger {
	if DefaultLogger == nil {
		return slog.With(args...)
	}
	return DefaultLogger.With(args...)
}

func SetUploadId(ctx context.Context, uploadId string) context.Context {
	logger := slog.With("uploadId", uploadId)
	return context.WithValue(ctx, LoggerKey, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(LoggerKey).(*slog.Logger)
	if !ok {
		// Fallback to the default logger if no logger is found in the context
		return slog.Default()
	}
	return logger
}
