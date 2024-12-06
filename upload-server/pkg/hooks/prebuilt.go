package hooks

import (
	"github.com/tus/tusd/v2/pkg/handler"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
)

type PrebuiltHook struct {
	hookMapping map[tusHooks.HookType][]HookHandlerFunc
}

type HookHandlerFunc func(event handler.HookEvent, resp tusHooks.HookResponse) (tusHooks.HookResponse, error)

func (ph *PrebuiltHook) Setup() error {
	return nil
}

func (ph *PrebuiltHook) InvokeHook(req tusHooks.HookRequest) (res tusHooks.HookResponse, err error) {
	hookFuncs, ok := ph.hookMapping[req.Type]
	if !ok {
		// nothing registered
		return res, nil
	}

	resp := tusHooks.HookResponse{}
	for _, hf := range hookFuncs {
		resp, err = hf(req.Event, resp)

		// Return early if we got an error.
		if err != nil {
			return resp, err
		}

		// Return early if a middleware function set a response.
		if resp.HTTPResponse.StatusCode != 0 {
			return resp, nil
		}
	}

	return resp, nil
}

func (ph *PrebuiltHook) Register(t tusHooks.HookType, hookFuncs ...HookHandlerFunc) {
	if ph.hookMapping == nil {
		ph.hookMapping = map[tusHooks.HookType][]HookHandlerFunc{}
	}
	ph.hookMapping[t] = append(ph.hookMapping[t], hookFuncs...)
}
