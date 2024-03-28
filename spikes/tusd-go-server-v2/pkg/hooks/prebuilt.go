package hooks

import (
	"github.com/tus/tusd/v2/pkg/handler"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
)

type PrebuiltHook struct {
	hookMapping map[tusHooks.HookType]func(handler.HookEvent) (tusHooks.HookResponse, error)
}

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

func (ph *PrebuiltHook) Register(t tusHooks.HookType, hook func(handler.HookEvent) (tusHooks.HookResponse, error)) {
	if ph.hookMapping == nil {
		ph.hookMapping = map[tusHooks.HookType]func(handler.HookEvent) (tusHooks.HookResponse, error){}
	}
	// TODO: defensive programming check that hooktype is one of the expected values
	ph.hookMapping[t] = hook
}
