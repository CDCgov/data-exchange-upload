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
	// Create a new logger to reset the fields to ensure only one upload ID is logged
	logger := slog.New(slog.Default().Handler())
	logger = slog.New(logger.Handler())
	logger.Info("New logger created")
	logger = logger.With("uploadId", uploadId)
	logger.Info("Logger with upload ID set")
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
