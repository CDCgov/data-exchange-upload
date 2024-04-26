package storeaz

import (
	"errors"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
) // .import

var (
	tusPrefix                        = "Tus storage"
	routerPrefix                     = "Router storage"
	edavPrefix                       = "Edav storage"
	errStorageNameEmpty              = errors.New("error storage name from app config is empty")
	errStorageKeyEmpty               = errors.New("error storage key from app config is empty")
	errStorageContainerEndpointEmpty = errors.New("error storage container endpoint from app config is empty")
) // .var

// NewTusAzBlobClient returns a azure blob client
func NewBlobClient(conf appconfig.AzureStorageConfig) (*azblob.Client, error) {

	return newAzBlobClient(
		conf.StorageName,
		conf.StorageKey,
		conf.ContainerEndpoint)
} // .NewTusAzBlobClient

func NewContainerClient(conf appconfig.AzureStorageConfig, containerName string) (*container.Client, error) {
	return newAzContainerClient(
		conf.StorageName,
		conf.StorageKey,
		conf.ContainerEndpoint,
		containerName)
}

func newAzContainerClient(azStorageName, azStorageKey, azContainerEndpoint, azContainerName string) (*container.Client, error) {
	// check guard if names are not empty
	if len(strings.TrimSpace(azStorageName)) == 0 {
		return nil, errStorageNameEmpty
	} // .if

	// check guard if names are not empty
	if len(strings.TrimSpace(azStorageKey)) == 0 {
		return nil, errStorageKeyEmpty
	} // .if

	// check guard if names are not empty
	if len(strings.TrimSpace(azContainerEndpoint)) == 0 {
		return nil, errStorageContainerEndpointEmpty
	} // .if

	credential, err := azblob.NewSharedKeyCredential(azStorageName, azStorageKey)
	if err != nil {
		return nil, err
	} // .if

	client, err := azblob.NewClientWithSharedKeyCredential(azContainerEndpoint, credential, nil)
	if err != nil {
		return nil, err
	} // .if

	return client.ServiceClient().NewContainerClient(azContainerName), nil
}

// newAzBlobClient, method for returning azure blob client for a storage needed
func newAzBlobClient(azStorageName, azStorageKey, azContainerEndpoint string) (*azblob.Client, error) {

	// check guard if names are not empty
	if len(strings.TrimSpace(azStorageName)) == 0 {
		return nil, errStorageNameEmpty
	} // .if

	// check guard if names are not empty
	if len(strings.TrimSpace(azStorageKey)) == 0 {
		return nil, errStorageKeyEmpty
	} // .if

	// check guard if names are not empty
	if len(strings.TrimSpace(azContainerEndpoint)) == 0 {
		return nil, errStorageContainerEndpointEmpty
	} // .if

	// getting the client

	credential, err := azblob.NewSharedKeyCredential(azStorageName, azStorageKey)
	if err != nil {
		return nil, err
	} // .if

	client, err := azblob.NewClientWithSharedKeyCredential(azContainerEndpoint, credential, nil)
	if err != nil {
		return nil, err
	} // .if

	return client, nil
} // .newAzBlobClient
