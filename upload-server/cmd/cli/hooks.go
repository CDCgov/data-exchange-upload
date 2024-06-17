package cli

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/upload"
	"os"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	azureloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/azure"
	fileloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/file"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/hooks/file"
)

func GetHookHandler(ctx context.Context, appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	if Flags.FileHooksDir != "" {
		return &file.FileHook{
			Directory: Flags.FileHooksDir,
		}, nil
	}
	return PrebuiltHooks(ctx, appConfig)
}

func PrebuiltHooks(ctx context.Context, appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	metadata.Cache = &metadata.ConfigCache{
		Loader: &fileloader.FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	manifestValidator := metadata.SenderManifestVerification{
		Configs: metadata.Cache,
	}

	var metadataAppender metadata.Appender
	metadataAppender = &metadata.FileMetadataAppender{
		Path: appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix,
	}

	dexFileDeliverer, err := postprocessing.NewFileDeliverer(ctx, "dex", &appConfig)
	if err != nil {
		return nil, err
	}
	edavFileDeliverer, err := postprocessing.NewFileDeliverer(ctx, "edav", &appConfig)
	if err != nil {
		return nil, err
	}
	routingFileDeliverer, err := postprocessing.NewFileDeliverer(ctx, "routing", &appConfig)
	if err != nil {
		return nil, err
	}

	if appConfig.AzureConnection != nil {
		client, err := storeaz.NewBlobClient(*appConfig.AzureConnection)
		if err != nil {
			return nil, err
		}
		metadata.Cache.Loader = &azureloader.AzureConfigLoader{
			Client:        client,
			ContainerName: appConfig.AzureManifestConfigContainer,
		}

		tusContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
		if err != nil {
			return nil, err
		}

		metadataAppender = &metadata.AzureMetadataAppender{
			ContainerClient: tusContainerClient,
			TusPrefix:       appConfig.TusUploadPrefix,
		}

		dexAzureDeliverer, err := postprocessing.NewAzureDeliverer(ctx, "dex", &appConfig)
		if err != nil {
			logger.Error("failed to connect to dex deliverer target", "error", err.Error())
		} else {
			postprocessing.RegisterTarget("dex", dexAzureDeliverer)
			health.Register(dexAzureDeliverer)
		}
	} else {
		postprocessing.RegisterTarget("dex", dexFileDeliverer)
		health.Register(dexFileDeliverer)
	}

	if appConfig.EdavConnection != nil {
		edavAzureDeliverer, err := postprocessing.NewAzureDeliverer(ctx, "edav", &appConfig)
		if err != nil {
			logger.Error("failed to connect to edav deliverer target", "error", err.Error())
		} else {
			postprocessing.RegisterTarget("edav", edavAzureDeliverer)
			health.Register(edavAzureDeliverer)
		}
	} else {
		postprocessing.RegisterTarget("edav", edavFileDeliverer)
		health.Register(edavFileDeliverer)
	}

	if appConfig.RoutingConnection != nil {
		routingAzureDeliverer, err := postprocessing.NewAzureDeliverer(ctx, "routing", &appConfig)
		if err != nil {
			logger.Error("failed to connect to router deliverer target", "error", err.Error())
		} else {
			postprocessing.RegisterTarget("routing", routingAzureDeliverer)
			health.Register(routingAzureDeliverer)
		}
	} else {
		postprocessing.RegisterTarget("routing", routingFileDeliverer)
		health.Register(routingFileDeliverer)
	}

	handler.Register(tusHooks.HookPreCreate, metadata.WithUploadID, metadata.WithTimestamp, manifestValidator.Verify)
	handler.Register(tusHooks.HookPostReceive, upload.ReportUploadStatus)
	handler.Register(tusHooks.HookPostCreate, upload.ReportUploadStarted)
	// note that tus sends this to a potentially blocking channel.
	// however it immediately pulls from that channel in to a goroutine..so we're good
	handler.Register(tusHooks.HookPostFinish, upload.ReportUploadComplete, manifestValidator.Hydrate, metadataAppender.Append, postprocessing.RouteAndDeliverHook)

	return handler, nil
}
