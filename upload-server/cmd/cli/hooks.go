package cli

import (
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/upload"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/hooks/file"
)

func GetHookHandler(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	if Flags.FileHooksDir != "" {
		return &file.FileHook{
			Directory: Flags.FileHooksDir,
		}, nil
	}

	manifestValidator := metadata.SenderManifestVerification{
		Configs: metadata.Cache,
	}

	var metadataAppender metadata.Appender = &metadata.FileMetadataAppender{
		Path: appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix,
	}

	if appConfig.AzureConnection != nil {
		tusContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
		if err != nil {
			return nil, err
		}

		metadataAppender = &metadata.AzureMetadataAppender{
			ContainerClient: tusContainerClient,
			TusPrefix:       appConfig.TusUploadPrefix,
		}
	}

	return PrebuiltHooks(manifestValidator, metadataAppender)
}

func PrebuiltHooks(validator metadata.SenderManifestVerification, appender metadata.Appender) (tusHooks.HookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	handler.Register(tusHooks.HookPreCreate, metadata.WithPreCreateManifestTransforms, validator.Verify)
	handler.Register(tusHooks.HookPostCreate, upload.ReportUploadStarted)
	handler.Register(tusHooks.HookPostReceive, upload.ReportUploadStatus)
	handler.Register(tusHooks.HookPreFinish, validator.Hydrate, appender.Append)
	// note that tus sends this to a potentially blocking channel.
	// however it immediately pulls from that channel in to a goroutine..so we're good
	handler.Register(tusHooks.HookPostFinish, upload.ReportUploadComplete, postprocessing.RouteAndDeliverHook())

	return handler, nil
}
