package azureinspector

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
)

type AzureUploadInspector struct {
	TusContainerClient *container.Client
	TusPrefix          string
}

func NewAzureUploadInspector(containerClient *container.Client, tusPrefix string) *AzureUploadInspector {
	return &AzureUploadInspector{
		TusContainerClient: containerClient,
		TusPrefix:          tusPrefix,
	}
}

func (aui *AzureUploadInspector) InspectInfoFile(c context.Context, id string) (map[string]any, error) {
	infoBlobClient := aui.TusContainerClient.NewBlobClient(aui.TusPrefix + "/" + id + ".info")

	// Download info file from blob client.
	downloadResponse, err := infoBlobClient.DownloadStream(c, nil)
	if err != nil {
		azErr, ok := err.(*azcore.ResponseError)
		if ok && azErr.StatusCode == http.StatusNotFound {
			return nil, errors.Join(err, info.ErrNotFound)
		}
		return nil, err
	}

	fileBytes, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		return nil, err
	}

	// Deserialize to hash map.
	jsonMap := &info.InfoFileData{}
	if err := json.Unmarshal(fileBytes, jsonMap); err != nil {
		return nil, err
	}

	return jsonMap.MetaData, nil
}

func (aui *AzureUploadInspector) InspectUploadedFile(c context.Context, id string) (map[string]any, error) {
	uploadBlobClient := aui.TusContainerClient.NewBlobClient(aui.TusPrefix + "/" + id)
	propertiesResponse, err := uploadBlobClient.GetProperties(c, nil)
	if err != nil {
		return nil, errors.Join(err, info.ErrNotFound)
	}

	uploadedFileInfo := map[string]any{
		"updated_at": propertiesResponse.LastModified.Format(time.RFC3339Nano),
		"size_bytes": propertiesResponse.ContentLength,
	}

	return uploadedFileInfo, nil
}
