package azureinspector

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"

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
	filename := filepath.Join(aui.TusPrefix, id+".info")
	infoBlobClient := aui.TusContainerClient.NewBlobClient(filename)

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
	filename := filepath.Join(aui.TusPrefix, id)
	uploadBlobClient := aui.TusContainerClient.NewBlobClient(filename)
	propertiesResponse, err := uploadBlobClient.GetProperties(c, nil)
	if err != nil {
		return nil, errors.Join(err, info.ErrNotFound)
	}

	uploadedFileInfo := map[string]any{
		"updated_at": propertiesResponse.LastModified,
		"size_bytes": propertiesResponse.ContentLength,
	}

	return uploadedFileInfo, nil
}

func (aui *AzureUploadInspector) InspectFileStatus(ctx context.Context, id string) (*info.DeliveryStatus, error) {
	//TODO implement me
	panic("implement me")
}
