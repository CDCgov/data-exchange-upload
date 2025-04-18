package logutil

import (
	"context"
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
)

func NewUploadIdLogger(event *handler.HookEvent, uploadId string) *slog.Logger {
	// Create a new context for the event to avoid contamination
	event.Context = context.Background()
	event.Context = sloger.SetUploadId(event.Context, uploadId)
	logger := sloger.GetLogger(event.Context)
	logger.Info("New logger with upload ID")

	return logger
}
