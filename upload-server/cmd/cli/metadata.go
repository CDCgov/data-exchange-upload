package cli

import (
	"context"
	"errors"
	"os"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders"
	azureloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/azure"
	fileloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/file"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
)

func InitConfigCache(ctx context.Context, appConfig appconfig.AppConfig) error {
	metadata.Cache = &metadata.ConfigCache{
		Loader: &fileloader.FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	if appConfig.AzureConnection != nil && appConfig.S3Connection != nil {
		return errors.New("cannot load metadata config from multiple locations")
	}

	if appConfig.AzureConnection != nil && appConfig.AzureManifestConfigContainer != "" {
		client, err := storeaz.NewBlobClient(appConfig.AzureConnection.Credentials())
		if err != nil {
			return err
		}
		metadata.Cache.Loader = &azureloader.AzureConfigLoader{
			Client:        client,
			ContainerName: appConfig.AzureManifestConfigContainer,
		}
	}

	if appConfig.S3Connection != nil && appConfig.S3ManifestConfigFolder != "" {
		client, err := stores3.NewWithEndpoint(ctx, appConfig.S3Connection.Endpoint)
		bucket := appConfig.S3Connection.BucketName
		if appConfig.S3ManifestConfigBucket != "" {
			bucket = appConfig.S3ManifestConfigBucket
		}
		if err != nil {
			return err
		}
		metadata.Cache.Loader = &loaders.S3ConfigLoader{
			Client:     client,
			BucketName: bucket,
			Folder:     appConfig.S3ManifestConfigFolder,
		}
	}

	return nil
}
