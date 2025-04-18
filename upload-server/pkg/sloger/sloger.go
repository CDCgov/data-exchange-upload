package sloger

import (
	"context"
	"log/slog"

	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
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
	logger := slog.With("uploadId", uploadId)
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

func WithUploadIdContext(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	event.Context = SetUploadId(event.Context, event.Upload.ID)
	return resp, nil
}

func Debug(ctx context.Context, msg string, args ...any) {
	GetLogger(ctx).Debug(msg, args...)
}

func Info(ctx context.Context, msg string, args ...any) {
	GetLogger(ctx).Info(msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	GetLogger(ctx).Error(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	GetLogger(ctx).Warn(msg, args...)
}
