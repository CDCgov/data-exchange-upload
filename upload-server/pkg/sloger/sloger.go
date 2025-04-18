package sloger

import (
	"context"
	"log/slog"
)

type contextKey string

const loggerKey contextKey = "logger"

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
	logger := GetLogger(ctx)
	logger = logger.With("uploadId", uploadId)
	return context.WithValue(ctx, loggerKey, logger)
}

func GetLogger(ctx context.Context) *slog.Logger {
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok {
		// Fallback to the default logger if no logger is found in the context
		return slog.Default()
	}
	return logger
}
