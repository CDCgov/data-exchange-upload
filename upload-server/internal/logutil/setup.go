package logutil

import (
	"context"
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
)

func SetupLogger(event *handler.HookEvent, uploadId string) *slog.Logger {
	// Create a new context for the event to avoid contamination
	event.Context = context.WithValue(context.Background(), "parentContext", event.Context)
	event.Context = sloger.SetUploadId(event.Context, uploadId)
	logger := sloger.GetLogger(event.Context)

	// Log the setup
	logger.Info("Logger setup with upload ID", "uploadId", uploadId)

	return logger
}
