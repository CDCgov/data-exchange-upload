package storeaz

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
) // .import

var (
	tusPrefix                        = "Tus"
	routerPrefix                     = "Router"
	edavPrefix                       = "Edav"
	errStorageNameEmpty              = "error storage name from app config is empty"
	errStorageKeyEmpty               = "error storage key from app config is empty"
	errStorageContainerEndpointEmpty = "error storage container endpoint from app config is empty"
) // .var

// NewTusAzBlobClient returns a azure blob client
func NewTusAzBlobClient(appConfig appconfig.AppConfig) (*azblob.Client, error) {

	return newAzBlobClient(tusPrefix, appConfig.TusAzStorageConfig.AzContainerName, appConfig.TusAzStorageConfig.AzStorageKey, appConfig.TusAzStorageConfig.AzContainerEndpoint)
} // .NewTusAzBlobClient

// NewRouterAzBlobClient returns a azure blob client
func NewRouterAzBlobClient(appConfig appconfig.AppConfig) (*azblob.Client, error) {

	return newAzBlobClient(routerPrefix, appConfig.RouterAzStorageConfig.AzContainerName, appConfig.RouterAzStorageConfig.AzStorageKey, appConfig.RouterAzStorageConfig.AzContainerEndpoint)
} // .NewRouterAzBlobClient

// NewEdavAzBlobClient returns a azure blob client
func NewEdavAzBlobClient(appConfig appconfig.AppConfig) (*azblob.Client, error) {

	return newAzBlobClient(edavPrefix, appConfig.EdavAzStorageConfig.AzContainerName, appConfig.EdavAzStorageConfig.AzStorageKey, appConfig.EdavAzStorageConfig.AzContainerEndpoint)
} // .NewEdavAzBlobClient

// newAzBlobClient, method for returning azure blob client for a storage needed
func newAzBlobClient(prefix, azStorageName, azStorageKey, azContainerEndpoint string) (*azblob.Client, error) {

	// check guard if names are not empty
	if len(strings.TrimSpace(azStorageName)) == 0 {
		return nil, NewError(prefix, errStorageNameEmpty)
	} // .if

	// check guard if names are not empty
	if len(strings.TrimSpace(azStorageKey)) == 0 {
		return nil, NewError(prefix, errStorageKeyEmpty)
	} // .if

	// check guard if names are not empty
	if len(strings.TrimSpace(azContainerEndpoint)) == 0 {
		return nil, NewError(prefix, errStorageContainerEndpointEmpty)
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
