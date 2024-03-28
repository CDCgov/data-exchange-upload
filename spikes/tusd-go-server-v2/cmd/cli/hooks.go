package cli

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/hooks"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/handler"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
)

func GetHookHandler() tusHooks.HookHandler {
	//TODO This can make decisions based on flags as the tusd implemenation does
	return PrebuiltHooks()
}

func HookHandlerFunc(f func(handler.HookEvent) (handler.HTTPResponse, handler.FileInfoChanges, error)) func(handler.HookEvent) (tusHooks.HookResponse, error) {
	return func(e handler.HookEvent) (res tusHooks.HookResponse, err error) {
		resp, changes, err := f(e)
		res.HTTPResponse = resp
		res.ChangeFileInfo = changes
		return res, err
	}
}

func PrebuiltHooks() tusHooks.HookHandler {
	handler := &prebuilthooks.PrebuiltHook{}
	handler.Register(tusHooks.HookPreCreate, HookHandlerFunc(hooks.CheckManifestV1()))
	handler.Register(tusHooks.HookPostFinish, hooks.LocalPostProcess)
	return handler
}
