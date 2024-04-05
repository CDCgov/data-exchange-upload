package cli

import (
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/handler"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/hooks/file"
)

func GetHookHandler() tusHooks.HookHandler {
	if Flags.FileHooksDir != "" {
		return &file.FileHook{
			Directory: Flags.FileHooksDir,
		}
	}
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
	return handler
}
