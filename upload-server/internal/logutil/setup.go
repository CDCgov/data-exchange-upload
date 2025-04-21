package logutil

import (
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func WithUploadIdLogger(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	tuid, err := metadata.GetUploadId(*event, resp)
	if err != nil {
		return resp, err
	}
	ctx, logger := sloger.SetInContext(event.Context, "uploadId", tuid)
	event.Context = ctx
	logger.Info("Logger setup with upload ID")
	return resp, nil
}
