package postprocessing

import (
	evt "github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func RouteAndDeliverHook(p evt.Publisher[evt.FileReady]) func(handler.HookEvent, hooks.HookResponse) (hooks.HookResponse, error) {
	return func(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
		id := event.Upload.ID
		var targets []string
		meta := event.Upload.MetaData
		if resp.ChangeFileInfo.MetaData != nil {
			meta = resp.ChangeFileInfo.MetaData
		}

		// Load config from metadata.
		path, err := metadata.GetConfigIdentifierByVersion(meta)
		if err != nil {
			return resp, err
		}
		config, err := metadata.Cache.GetConfig(event.Context, path)
		if err != nil {
			return resp, err
		}
		targets = append(targets, config.Copy.Targets...)

		for _, target := range targets {
			e := evt.NewFileReadyEvent(id, meta, target)
			err := p.Publish(event.Context, e)
			if err != nil {
				return resp, err
			}
			logger.Info("published event", "event", e)
		}
		return resp, nil
	}
}
