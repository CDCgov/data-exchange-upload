package postprocessing

import (
	"errors"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"

	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func RouteAndDeliverHook(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	id := event.Upload.ID
	//put a message on a queue system
	//the message should have the tuid and the manifest
	//a deliverer writes the file with the manifest as the files metadata

	// should eventually take a tuid and that's it
	// why don't we just do this n times, once internal, once to edav, once to routing (whatever the number of targets is?)
	targets := []string{
		"dex",
	}
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

	var errs error
	for _, target := range targets {
		// fan out command
		// send an event for each thing to be copied
		errs = errors.Join(errs, Deliver(id, meta, target))
	}
	return resp, errs
}
