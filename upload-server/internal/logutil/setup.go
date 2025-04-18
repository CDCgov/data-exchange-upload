package logutil

import (
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func SetupLogger(event *handler.HookEvent, uploadId string) *slog.Logger {
	event.Context = sloger.SetUploadId(event.Context, uploadId)
	logger := sloger.GetLogger(event.Context)
	logger.Info("Logger setup with upload ID", uploadId)
	return logger
}

func WithUploadIdLogger(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	event.Context = sloger.SetUploadId(event.Context, event.Upload.ID)
	slog.SetDefault(sloger.GetLogger(event.Context))
	slog.Info("Logger setup with upload ID")
	return resp, nil
}
