package cli

import (
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/logutil"
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

func GetHookHandler(appConfig appconfig.AppConfig) (RegisterableHookHandler, error) {

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

	if appConfig.S3Connection != nil {
		// Tus S3 store automatically appends metadata to the S3 object
		metadataAppender = &metadata.NoopAppender{}
	}

	return PrebuiltHooks(manifestValidator, metadataAppender)
}

func PrebuiltHooks(validator metadata.SenderManifestVerification, appender metadata.Appender) (RegisterableHookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	handler.Register(tusHooks.HookPreCreate, metadata.WithUploadId, logutil.WithUploadIdLogger, metadata.WithPreCreateManifestTransforms, validator.Verify)
	handler.Register(tusHooks.HookPostCreate, logutil.WithUploadIdLogger, upload.ReportUploadStarted)
	handler.Register(tusHooks.HookPostReceive, logutil.WithUploadIdLogger, upload.ReportUploadStatus)
	handler.Register(tusHooks.HookPreFinish, logutil.WithUploadIdLogger, appender.Append)
	// note that tus sends this to a potentially blocking channel.
	// however it immediately pulls from that channel in to a goroutine..so we're good

	handler.Register(tusHooks.HookPostFinish, logutil.WithUploadIdLogger, upload.ReportUploadComplete, postprocessing.RouteAndDeliverHook())

	return handler, nil
}
