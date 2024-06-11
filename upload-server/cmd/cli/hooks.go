package cli

import (
	"context"
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

func GetHookHandler(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	if Flags.FileHooksDir != "" {
		return &file.FileHook{
			Directory: Flags.FileHooksDir,
		}, nil
	}
	return PrebuiltHooks(appConfig)
}

func PrebuiltHooks(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
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

	postprocessing.RegisterTarget("dex", &postprocessing.FileDeliverer{
		ToPath: appConfig.LocalDEXFolder,
		From:   os.DirFS(appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix),
	})

	postprocessing.RegisterTarget("edav", &postprocessing.FileDeliverer{
		ToPath: appConfig.LocalEDAVFolder,
		From:   os.DirFS(appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix),
	})

	postprocessing.RegisterTarget("routing", &postprocessing.FileDeliverer{
		ToPath: appConfig.LocalROUTINGFolder,
		From:   os.DirFS(appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix),
	})

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
		dexCheckpointContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.DexCheckpointContainer)
		if err != nil {
			return nil, err
		}
		edavCheckpointContainerClient, err := storeaz.NewContainerClient(*appConfig.EdavConnection, appConfig.EdavCheckpointContainer)
		if err != nil {
			return nil, err
		}
		routingCheckpointContainerClient, err := storeaz.NewContainerClient(*appConfig.RoutingConnection, appConfig.RoutingCheckpointContainer)
		if err != nil {
			return nil, err
		}

		ctx := context.Background()
		err = storeaz.CreateContainerIfNotExists(ctx, dexCheckpointContainerClient)
		if err != nil {
			return nil, err
		}
		err = storeaz.CreateContainerIfNotExists(ctx, edavCheckpointContainerClient)
		if err != nil {
			return nil, err
		}
		err = storeaz.CreateContainerIfNotExists(ctx, routingCheckpointContainerClient)
		if err != nil {
			return nil, err
		}

		postprocessing.RegisterTarget("dex", &postprocessing.AzureDeliverer{
			FromContainerClient: tusContainerClient,
			ToContainerClient:   dexCheckpointContainerClient,
			TusPrefix:           appConfig.TusUploadPrefix,
			Target:              "dex",
		})
		postprocessing.RegisterTarget("edav", &postprocessing.AzureDeliverer{
			FromContainerClient: tusContainerClient,
			ToContainerClient:   edavCheckpointContainerClient,
			TusPrefix:           appConfig.TusUploadPrefix,
			Target:              "edav",
		})
		postprocessing.RegisterTarget("routing", &postprocessing.AzureDeliverer{
			FromContainerClient: tusContainerClient,
			ToContainerClient:   routingCheckpointContainerClient,
			TusPrefix:           appConfig.TusUploadPrefix,
			Target:              "routing",
		})

		metadataAppender = &metadata.AzureMetadataAppender{
			ContainerClient: tusContainerClient,
			TusPrefix:       appConfig.TusUploadPrefix,
		}
	}

	handler.Register(tusHooks.HookPreCreate, metadata.WithUploadID, metadata.WithTimestamp, manifestValidator.Verify)
	handler.Register(tusHooks.HookPostReceive, upload.ReportUploadStatus)
	handler.Register(tusHooks.HookPostCreate, upload.ReportUploadStarted)
	// note that tus sends this to a potentially blocking channel.
	// however it immediately pulls from that channel in to a goroutine..so we're good
	handler.Register(tusHooks.HookPostFinish, upload.ReportUploadComplete, manifestValidator.Hydrate, metadataAppender.Append, postprocessing.RouteAndDeliverHook)

	return handler, nil
}
