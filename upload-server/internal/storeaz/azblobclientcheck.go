package storeaz

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/tus/tusd/v2/pkg/azurestore"
) // .import

type AzureBlobHealthCheck struct {
	client *azblob.Client
}

func NewAzureHealthCheck(conf *azurestore.AzConfig) (*AzureBlobHealthCheck, error) {
	client, err := newAzBlobClient(tusPrefix, conf.AccountName, conf.AccountKey, conf.Endpoint, conf.ContainerName)
	if err != nil {
		return nil, err
	}
	return &AzureBlobHealthCheck{
		client: client,
	}, nil
}

func (c *AzureBlobHealthCheck) Health(ctx context.Context) models.ServiceHealthResp {
	return checkAzBlobClient(tusPrefix, c.client)
}

// checkAzBlobClient, method for checking still valid and working the azure blob client for a storage
func checkAzBlobClient(prefix string, client *azblob.Client) models.ServiceHealthResp {

	var shr models.ServiceHealthResp
	shr.Service = prefix

	// guard client is null
	if client == nil {
		shr.Service = models.STATUS_DOWN
		shr.HealthIssue = models.AZ_BLOB_CLIENT_NA
		return shr
	} // .if

	// test if the client is good
	_, err := client.CreateContainer(context.TODO(), models.AZ_TEST_CONTAINER_NAME, nil)

	// check to see if error is blob does exists which means client is ok
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		if responseErr.ErrorCode == string(bloberror.ContainerAlreadyExists) {
			// connection ok
			shr.Status = models.STATUS_UP
			shr.HealthIssue = models.HEALTH_ISSUE_NONE
			return shr
		} // .if
	} // .if

	if err != nil {
		shr.Status = models.STATUS_DOWN
		shr.HealthIssue = err.Error()
		return shr
	} // .if

	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
} // .checkAzBlobClient
