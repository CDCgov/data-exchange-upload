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
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/handler"
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

func PrebuiltHooks(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	preCreateHook := metadata.SenderManifestVerification{
		Loader: &FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	if appConfig.AzureConnection != nil {
		client, err := storeaz.NewBlobClient(*appConfig.AzureConnection)
		if err != nil {
			return nil, err
		}
		preCreateHook.Loader = &AzureConfigLoader{
			Client:        client,
			ContainerName: appConfig.AzureManifestConfigContainer,
		}
	}

	handler.Register(tusHooks.HookPreCreate, preCreateHook.Verify)
	return handler, nil
}
