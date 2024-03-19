package storeaz

import (
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
) // .import

// NewAzBlobClient returns a azure blob client
func NewAzBlobClient(appConfig appconfig.AppConfig) (*azblob.Client, error) {

	credential, err := azblob.NewSharedKeyCredential(appConfig.AzStorageName, appConfig.AzStorageKey)
	if err != nil {
		return nil, err
	} // .if

	client, err := azblob.NewClientWithSharedKeyCredential(appConfig.AzContainerEndpoint, credential, nil)
	if err != nil {
		return nil, err
	} // .if

	return client, nil
} // .NewAzBlobClient
