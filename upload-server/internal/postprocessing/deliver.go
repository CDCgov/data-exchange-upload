package postprocessing

import (
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func Deliver(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	// root -> ""
	// date -> default pattern
	// "" -> ""
	return resp, nil
}
