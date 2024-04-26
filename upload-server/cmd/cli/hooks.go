package cli

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/hooks/file"
)

func GetHookHandler(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	if Flags.FileHooksDir != "" {
		return &file.FileHook{
			Directory: Flags.FileHooksDir,
		}, nil
	}
	return PrebuiltHooks(appConfig)
}

func HookHandlerFunc(f func(handler.HookEvent) (handler.HTTPResponse, handler.FileInfoChanges, error)) func(handler.HookEvent) (tusHooks.HookResponse, error) {
	return func(e handler.HookEvent) (res tusHooks.HookResponse, err error) {
		resp, changes, err := f(e)
		res.HTTPResponse = resp
		res.ChangeFileInfo = changes
		return res, err
	}
}

type FileConfigLoader struct {
	FileSystem fs.FS
}

func (l *FileConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {

	file, err := l.FileSystem.Open(path)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}

type AzureConfigLoader struct {
	Client        *azblob.Client
	ContainerName string
}

func (l *AzureConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {
	downloadResponse, err := l.Client.DownloadStream(ctx, l.ContainerName, path, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if respErr.StatusCode == http.StatusNotFound {
				return nil, os.ErrNotExist
			}
		}
		return nil, err
	}

	return io.ReadAll(downloadResponse.Body)
}

// TODO: Combine somewhere for all hooks needing ConfigLoader
//
//	       like preCreate?
//		      Is currently a duplicate for  metadata.SenderManifestVerification
type HookConfigLoader struct {
	Loader validation.ConfigLoader
}

// TODO: Relocate in to maybe internal/hooks or internal/upload-status ?
func (v *HookConfigLoader) PostReceive(event handler.HookEvent) (hooks.HookResponse, error) {
	resp := hooks.HookResponse{}

	// Get values from event
	uploadId := event.Upload.ID
	uploadSize := event.Upload.Size
	uploadOffset := event.Upload.Offset
	manifest := event.Upload.MetaData

	logger.Info(
		"[PostReceive]: event.Upload values",
		" manifest: ", manifest,
		" uploadId: ", uploadId,
		" uploadSize: ", uploadSize,
		" uploadOffset: ", uploadOffset,
	)

	// TODO: Add shell script logic here.

	// TODO: Covert Python post_receive_bin.py starting here...

	return resp, nil

}

func PrebuiltHooks(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	preCreateHook := metadata.SenderManifestVerification{
		Loader: &FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	postReceiveHook := HookConfigLoader{
		Loader: &FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	if appConfig.DexAzUploadConfig != nil {
		client, err := storeaz.NewBlobClient(*appConfig.DexAzUploadConfig)
		if err != nil {
			return nil, err
		}
		preCreateHook.Loader = &AzureConfigLoader{
			Client:        client,
			ContainerName: appConfig.DexAzUploadConfig.AzContainerName,
		}
	}

	handler.Register(tusHooks.HookPreCreate, preCreateHook.Verify)
	handler.Register(tusHooks.HookPostReceive, postReceiveHook.PostReceive)
	return handler, nil
}
