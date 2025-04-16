package logutil

import (
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func WithLoggerSetup(event *handler.HookEvent, resp hooks.HookResponse) (*slog.Logger, hooks.HookResponse) {
	if resp.ChangeFileInfo.ID != "" {
		event.Context = sloger.SetUploadId(event.Context, resp.ChangeFileInfo.ID)
	} else if event.Upload.ID != "" {
		event.Context = sloger.SetUploadId(event.Context, event.Upload.ID)
	} else {
		logger := sloger.GetLogger(event.Context)
		logger.Error("upload ID is not set")
		return logger, resp
	}
	logger := sloger.GetLogger(event.Context)

	logger.Info("Logger setup with upload ID")

	return logger, resp
}
