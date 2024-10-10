package storeaz

import (
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
) // .import

var (
	tusPrefix                        = "Tus storage"
	errStorageNameEmpty              = errors.New("error storage name from app config is empty")
	errStorageKeyEmpty               = errors.New("error storage key from app config is empty")
	errStorageContainerEndpointEmpty = errors.New("error storage container endpoint from app config is empty")
) // .var

// NewTusAzBlobClient returns a azure blob client
func NewBlobClient(conf appconfig.AzureStorageConfig) (*azblob.Client, error) {
	if canUseStorageKey(conf) {
		return newAzBlobClient(
			conf.StorageName,
			conf.StorageKey,
			conf.ContainerEndpoint)
	}

	if canUseServicePrinciple(conf) {
		return newAzBlobClientByServicePrinciple(conf)
	}

	return nil, errors.New("not enough information given to connect to storage account " + conf.StorageName)
} // .NewTusAzBlobClient

func NewContainerClient(conf appconfig.AzureStorageConfig, containerName string) (*container.Client, error) {
	if canUseStorageKey(conf) {
		return newAzContainerClient(
			conf.StorageName,
			conf.StorageKey,
			conf.ContainerEndpoint,
			containerName)
	}

	if canUseServicePrinciple(conf) {
		return newContainerClientByServicePrinciple(conf, containerName)
	}

	return nil, errors.New(fmt.Sprintf("not enough information given to connect to account %s and container %s", conf.ContainerEndpoint, containerName))
}

func newContainerClientByServicePrinciple(conf appconfig.AzureStorageConfig, containerName string) (*container.Client, error) {
	cred, err := azidentity.NewClientSecretCredential(conf.TenantId, conf.ClientId, conf.ClientSecret, nil)
	if err != nil {
		return nil, err
	}
	client, err := azblob.NewClient(conf.ContainerEndpoint, cred, nil)

	return client.ServiceClient().NewContainerClient(containerName), nil
}

func newAzContainerClient(azStorageName, azStorageKey, azContainerEndpoint, azContainerName string) (*container.Client, error) {
	client, err := newAzBlobClient(azStorageName, azStorageKey, azContainerEndpoint)
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

func newAzBlobClientByServicePrinciple(conf appconfig.AzureStorageConfig) (*azblob.Client, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	client, err := azblob.NewClient(conf.ContainerEndpoint, cred, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func canUseStorageKey(conf appconfig.AzureStorageConfig) bool {
	return conf.StorageKey != ""
}

func canUseServicePrinciple(conf appconfig.AzureStorageConfig) bool {
	return conf.TenantId != "" && conf.ClientId != "" && conf.ClientSecret != ""
}
