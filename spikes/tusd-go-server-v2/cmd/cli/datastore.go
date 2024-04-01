package cli

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlertusd"
	"github.com/tus/tusd/v2/pkg/azurestore"
	"github.com/tus/tusd/v2/pkg/filestore"
)

func CreateDataStore(appConfig appconfig.AppConfig) (handlertusd.Store, error) {
	// ------------------------------------------------------------------
	// Load Az dependencies, needed for the DEX handler paths
	// ------------------------------------------------------------------
	if appConfig.TusAzStorageConfig.AzContainerName != "" {
		if err := appConfig.TusAzStorageConfig.Check(); err != nil {
			return nil, err
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
			return nil, err
		} // azService

		return azurestore.New(azService), nil
		// store.ObjectPrefix = Flags.AzObjectPrefix
		// store.Container = appConfig.AzContainerName

		// TODO: set for azure
		// TODO: set for azure, Upload Locks: https://tus.github.io/tusd/advanced-topics/locks/
	} // .if
	// Create a new FileStore instance which is responsible for
	// storing the uploaded file on disk in the specified directory.
	// This path _must_ exist before tusd will store uploads in it.
	// If you want to save them on a different medium, for example
	// a remote FTP server, you can implement your own storage backend
	// by implementing the tusd.DataStore interface.
	return filestore.FileStore{
		Path: appConfig.LocalFolderUploadsTus,
	}, nil // .store
}
