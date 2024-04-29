package hooks

import (
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
)

type PrebuiltHook struct {
	hookMapping map[tusHooks.HookType]HookHandlerFunc
}

type HookHandlerFunc func(event handler.HookEvent) (hooks.HookResponse, error)

func (ph *PrebuiltHook) Setup() error {
	return nil
}

func (ph *PrebuiltHook) InvokeHook(req tusHooks.HookRequest) (res tusHooks.HookResponse, err error) {
	hook, ok := ph.hookMapping[req.Type]
	if !ok {
		// nothing registered
		return res, nil
	}
	return hook(req.Event)
}

func (ph *PrebuiltHook) Register(t tusHooks.HookType, hook HookHandlerFunc) {
	if ph.hookMapping == nil {
		ph.hookMapping = map[tusHooks.HookType]HookHandlerFunc{}
	}
	ph.hookMapping[t] = hook
}
