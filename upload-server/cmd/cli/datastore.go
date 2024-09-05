package cli

import (
	"context"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	"github.com/tus/tusd/v2/pkg/s3store"
	"os"
	"path/filepath"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/tus/tusd/v2/pkg/azurestore"
	"github.com/tus/tusd/v2/pkg/filestore"
)

func GetDataStore(ctx context.Context, appConfig appconfig.AppConfig) (handlertusd.Store, health.Checkable, error) {
	// ------------------------------------------------------------------
	// Load Az dependencies, needed for the DEX handler paths
	// ------------------------------------------------------------------
	if appConfig.AzureConnection != nil {
		if err := appConfig.AzureConnection.Check(); err != nil {
			return nil, nil, err
		}

		accountName := appConfig.AzureConnection.StorageName

		azureEndpoint := appConfig.AzureConnection.ContainerEndpoint
		logger.Info("Using Azure endpoint", "endpoint", azureEndpoint)

		azConfig := &azurestore.AzConfig{
			AccountName:   accountName,
			AccountKey:    appConfig.AzureConnection.StorageKey,
			ContainerName: appConfig.AzureUploadContainer,
			Endpoint:      azureEndpoint,
		} // .azConfig

		azService, err := azurestore.NewAzureService(azConfig)
		if err != nil {
			return nil, nil, err
		} // azService

		hc, err := storeaz.NewAzureHealthCheck(azConfig)

		if err != nil {
			return nil, nil, err
		} // azService

		store := azurestore.New(azService)
		store.ObjectPrefix = appConfig.TusUploadPrefix
		store.Container = appConfig.AzureUploadContainer
		return store, hc, nil
	} // .if

	if appConfig.S3Connection != nil {
		client, err := stores3.New(ctx, appConfig.S3Connection)
		if err != nil {
			return nil, nil, err
		}
		hc := &stores3.S3HealthCheck{
			Client:     client,
			BucketName: appConfig.S3Connection.BucketName,
		}
		store := s3store.New(appConfig.S3Connection.BucketName, client)
		store.ObjectPrefix = appConfig.TusUploadPrefix

		logger.Info("using S3 bucket", "bucket", appConfig.S3Connection.BucketName)
		return store, hc, nil
	}

	// Create a new FileStore instance which is responsible for
	// storing the uploaded file on disk in the specified directory.
	// This path _must_ exist before tusd will store uploads in it.
	// If you want to save them on a different medium, for example
	// a remote FTP server, you can implement your own storage backend
	// by implementing the tusd.DataStore interface.
	path := appConfig.LocalFolderUploadsTus
	path = filepath.Join(path, appConfig.TusUploadPrefix)

	return filestore.FileStore{
		Path: path,
	}, &FileStoreHealthCheck{path: appConfig.LocalFolderUploadsTus}, nil // .store
}

type FileStoreHealthCheck struct {
	path string
}

func (c *FileStoreHealthCheck) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "File Storage"
	info, err := os.Stat(c.path)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if !info.IsDir() {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = fmt.Sprintf("%s is not a directory", c.path)
		return rsp
	}
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}
