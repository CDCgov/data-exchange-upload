package storeaz

import (
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
) // .import

// NewTusAzBlobClient returns a azure blob client
func NewTusAzBlobClient(appConfig appconfig.AppConfig) (*azblob.Client, error) {

	credential, err := azblob.NewSharedKeyCredential(appConfig.TusAzStorageConfig.AzStorageName, appConfig.TusAzStorageConfig.AzStorageKey)
	if err != nil {
		return nil, err
	} // .if

	client, err := azblob.NewClientWithSharedKeyCredential(appConfig.TusAzStorageConfig.AzContainerEndpoint, credential, nil)
	if err != nil {
		return nil, err
	} // .if

	return client, nil
} // .NewTusAzBlobClient
