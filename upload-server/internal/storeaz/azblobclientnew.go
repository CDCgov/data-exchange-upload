package storeaz

import (
	"context"
	"errors"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
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
		conf.AzStorageName,
		conf.AzStorageKey,
		conf.AzContainerEndpoint,
		conf.AzContainerName)
} // .NewTusAzBlobClient

// newAzBlobClient, method for returning azure blob client for a storage needed
func newAzBlobClient(azStorageName, azStorageKey, azContainerEndpoint, azContainerName string) (*azblob.Client, error) {

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

	// test if the client is
	_, err = client.CreateContainer(context.TODO(), azContainerName, nil)

	// check to see if error is blob does exists which means client is ok
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		if responseErr.ErrorCode == string(bloberror.ContainerAlreadyExists) {
			// connection ok
			return client, nil
		} // .if
	} // .if

	if err != nil {
		return nil, err
	} // .if

	return client, nil
} // .newAzBlobClient
