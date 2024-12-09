package cli

import (
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/upload"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
)

type RegisterableHookHandler interface {
	tusHooks.HookHandler
	Register(t tusHooks.HookType, hookFuncs ...prebuilthooks.HookHandlerFunc)
}

func GetHookHandler(appConfig appconfig.AppConfig, eventPublisher event.Publisher[*event.FileReady]) (RegisterableHookHandler, error) {

	manifestValidator := metadata.SenderManifestVerification{
		Configs: metadata.Cache,
	}

	var metadataAppender metadata.Appender = &metadata.FileMetadataAppender{
		Path: appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix,
	}

	if appConfig.AzureConnection != nil {
		tusContainerClient, err := storeaz.NewContainerClient(appConfig.AzureConnection.Credentials(), appConfig.AzureUploadContainer)
		if err != nil {
			return nil, err
		}

		metadataAppender = &metadata.AzureMetadataAppender{
			ContainerClient: tusContainerClient,
			TusPrefix:       appConfig.TusUploadPrefix,
		}
	}

	return PrebuiltHooks(manifestValidator, metadataAppender, eventPublisher)
}

func PrebuiltHooks(validator metadata.SenderManifestVerification, appender metadata.Appender, eventPublisher event.Publisher[*event.FileReady]) (RegisterableHookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	handler.Register(tusHooks.HookPreCreate, metadata.WithPreCreateManifestTransforms, validator.Verify)
	handler.Register(tusHooks.HookPostCreate, upload.ReportUploadStarted)
	handler.Register(tusHooks.HookPostReceive, upload.ReportUploadStatus)
	handler.Register(tusHooks.HookPreFinish, appender.Append)
	// note that tus sends this to a potentially blocking channel.
	// however it immediately pulls from that channel in to a goroutine..so we're good

	handler.Register(tusHooks.HookPostFinish, upload.ReportUploadComplete, postprocessing.RouteAndDeliverHook(eventPublisher))

	return handler, nil
}
