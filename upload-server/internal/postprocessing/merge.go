package postprocessing

import (
	"context"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	evt "github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func RouteAndDeliverHook() func(handler.HookEvent, hooks.HookResponse) (hooks.HookResponse, error) {
	return func(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
		ctx := context.TODO()
		id := event.Upload.ID
		meta := event.Upload.MetaData
		if resp.ChangeFileInfo.MetaData != nil {
			meta = resp.ChangeFileInfo.MetaData
		}

		routeGroup, ok := delivery.FindGroupFromMetadata(meta)
		if !ok {
			return resp, fmt.Errorf("no routing group found for metadata %+v", meta)
		}

		for _, target := range routeGroup.TargetNames() {
			e := evt.NewFileReadyEvent(id, meta, target)
			err := evt.FileReadyPublisher.Publish(ctx, e)
			if err != nil {
				return resp, err
			}
			logger.Info("published event", "event", e)
		}
		return resp, nil
	}
}
