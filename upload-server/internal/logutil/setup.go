package logutil

import (
	"context"
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
)

func SetupLogger(event *handler.HookEvent, uploadId string) *slog.Logger {
	event.Context = sloger.SetUploadId(event.Context, uploadId)
	logger := sloger.GetLogger(event.Context)
	logger.Info("Logger setup with upload ID")
	return logger
}

func SetupLoggerWithContext(ctx context.Context, uploadId string) (context.Context, *slog.Logger) {
	ctx = sloger.SetUploadId(ctx, uploadId)
	logger := sloger.GetLogger(ctx)
	logger.Info("Logger setup with upload ID")
	return ctx, logger
}
