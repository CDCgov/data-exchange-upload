package storeaz

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/models"
) // .import

// CheckTusAzBlobClient returns a check on the azure blob client
func CheckTusAzBlobClient(client *azblob.Client) models.HealthServiceResp {

	return checkAzBlobClient(tusPrefix, client)
} // .CheckTusAzBlobClient

// CheckRouterAzBlobClient returns a check on the azure blob client
func CheckRouterAzBlobClient(client *azblob.Client) models.HealthServiceResp {

	return checkAzBlobClient(routerPrefix, client)
} // .CheckRouterAzBlobClient

// CheckEdavAzBlobClient returns a check on the azure blob client
func CheckEdavAzBlobClient(client *azblob.Client) models.HealthServiceResp {

	return checkAzBlobClient(edavPrefix, client)
} // .CheckEdavAzBlobClient

// checkAzBlobClient, method for checking still valid and working the azure blob client for a storage
func checkAzBlobClient(prefix string, client *azblob.Client) models.HealthServiceResp {

	var hsr models.HealthServiceResp
	hsr.Service = prefix

	// guard client is null
	if client == nil {
		hsr.Service = models.STATUS_DOWN
		hsr.HealthIssue = models.AZ_BLOB_CLIENT_NA
		return hsr
	} // .if

	// test if the client is good
	_, err := client.CreateContainer(context.TODO(), models.AZ_TEST_CONTAINER_NAME, nil)

	// check to see if error is blob does exists which means client is ok
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		if responseErr.ErrorCode == string(bloberror.ContainerAlreadyExists) {
			// connection ok
			hsr.Status = models.STATUS_UP
			hsr.HealthIssue = models.HEALTH_ISSUE_NONE
			return hsr
		} // .if
	} // .if

	if err != nil {
		hsr.Status = models.STATUS_DOWN
		hsr.HealthIssue = err.Error()
		return hsr
	} // .if

	hsr.Status = models.STATUS_UP
	hsr.HealthIssue = models.HEALTH_ISSUE_NONE
	return hsr
} // .checkAzBlobClient
