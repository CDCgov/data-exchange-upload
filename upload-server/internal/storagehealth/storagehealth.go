package storagehealth

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/data-exchange-upload/upload-server/internal/models"
)

// CheckBlobStorageHealth checks the health of Azure Blob Storage
func CheckBlobStorageHealth(ctx context.Context, storage string) models.ServiceHealthResp {
	var containerName, accountName, accountKey, connectionString string
	var blobServiceClient *azblob.ServiceClient
	var err error

	switch storage {
	case "EDAV storage":
		containerName = "dextesting-testevent1"
		accountName = os.Getenv("EDAV_AZURE_STORAGE_ACCOUNT_NAME")
		cred, _ := azidentity.NewDefaultAzureCredential(nil)
		blobServiceClient, err = azblob.NewServiceClient(fmt.Sprintf("https://%s.blob.core.windows.net", accountName), cred, nil)
	case "Routing Blob Container":
		containerName = "test-routing"
		accountName = os.Getenv("ROUTING_STORAGE_ACCOUNT_NAME")
		accountKey = os.Getenv("ROUTING_STORAGE_ACCOUNT_KEY")
		connectionString = fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=core.windows.net", accountName, accountKey)
		blobServiceClient, err = azblob.NewServiceClientFromConnectionString(connectionString, nil)

	}

	if err != nil {
		return models.ServiceHealthResp{
			Service:     storage,
			Status:      models.STATUS_DOWN,
			HealthIssue: fmt.Sprintf("Failed to create blob service client: %v", err),
		}
	}

	// Log the operation
	fmt.Printf("Checking health for destination: %s\n", storage)

	// Simulate checking the container's existence or other operation that confirms the health
	containerClient := blobServiceClient.NewContainerClient(containerName)
	_, err = containerClient.GetProperties(ctx, nil)
	if err != nil {
		return models.ServiceHealthResp{
			Service:     storage,
			Status:      models.STATUS_DOWN,
			HealthIssue: fmt.Sprintf("Health check failed: %v", err),
		}
	}

	return models.ServiceHealthResp{
		Service:     storage,
		Status:      models.STATUS_UP,
		HealthIssue: models.HEALTH_ISSUE_NONE,
	}
}
