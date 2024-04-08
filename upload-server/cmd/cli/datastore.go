package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/tus/tusd/v2/pkg/azurestore"
	"github.com/tus/tusd/v2/pkg/filestore"
)

func CreateDataStore(appConfig appconfig.AppConfig) (handlertusd.Store, health.Checkable, error) {
	// ------------------------------------------------------------------
	// Load Az dependencies, needed for the DEX handler paths
	// ------------------------------------------------------------------
	if appConfig.TusAzStorageConfig != nil && appConfig.TusAzStorageConfig.AzContainerName != "" {
		if err := appConfig.TusAzStorageConfig.Check(); err != nil {
			return nil, nil, err
		}

		azConfig := &azurestore.AzConfig{
			AccountName:         appConfig.TusAzStorageConfig.AzStorageName,
			AccountKey:          appConfig.TusAzStorageConfig.AzStorageKey,
			ContainerName:       appConfig.TusAzStorageConfig.AzContainerName,
			ContainerAccessType: appConfig.TusAzStorageConfig.AzContainerAccessType,
			// BlobAccessTier:      Flags.AzBlobAccessTier,
			Endpoint: appConfig.TusAzStorageConfig.AzContainerEndpoint,
		} // .azConfig

		azService, err := azurestore.NewAzureService(azConfig)
		if err != nil {
			return nil, nil, err
		} // azService

		hc, err := storeaz.NewAzureHealthCheck(azConfig)

		if err != nil {
			return nil, nil, err
		} // azService

		return azurestore.New(azService), hc, nil
	} // .if

	// Create a new FileStore instance which is responsible for
	// storing the uploaded file on disk in the specified directory.
	// This path _must_ exist before tusd will store uploads in it.
	// If you want to save them on a different medium, for example
	// a remote FTP server, you can implement your own storage backend
	// by implementing the tusd.DataStore interface.
	return filestore.FileStore{
		Path: appConfig.LocalFolderUploadsTus,
	}, &FileStoreHealthCheck{path: appConfig.LocalFolderUploadsTus}, nil // .store
}

type FileStoreHealthCheck struct {
	path string
}

func (c *FileStoreHealthCheck) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "File Storage"
	info, err := os.Stat(c.path)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
	}
	if !info.IsDir() {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = fmt.Sprintf("%s is not a directory", c.path)
	}
	rsp.Status = models.STATUS_UP
	return rsp
}
