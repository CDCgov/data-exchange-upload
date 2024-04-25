package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
)

var (
	ErrNotFound = errors.New("Expected file not found")
)

type UploadInspecter interface {
	InspectInfoFile(c context.Context, id string) (map[string]any, error)
	InspectUploadedFile(c context.Context, id string) (map[string]any, error)
}

func NewFileSystemUploadInspector(baseDir string) *FileSystemUploadInspector {
	return &FileSystemUploadInspector{
		BaseDir: baseDir,
	}
}

type FileSystemUploadInspector struct {
	BaseDir string
}

func NewAzureUploadInspector(containerClient *container.Client, tusDir string) *AzureUploadInspector {
	return &AzureUploadInspector{
		TusContainerClient: containerClient,
		TusDir:             tusDir,
	}
}

type AzureUploadInspector struct {
	TusContainerClient *container.Client
	TusDir             string
}

type InfoFileData struct {
	MetaData map[string]any `json:"MetaData"`
}

func (fsui *FileSystemUploadInspector) InspectInfoFile(c context.Context, id string) (map[string]any, error) {
	// First, read in the id + .info file.
	//TODO make this stronger
	infoFilename := fsui.BaseDir + "/" + id + ".info"
	fileBytes, err := os.ReadFile(infoFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.Join(err, ErrNotFound)
		}
		return nil, err
	}

	// Deserialize to hash map.
	jsonMap := &InfoFileData{}
	if err := json.Unmarshal(fileBytes, jsonMap); err != nil {
		return nil, err
	}

	return jsonMap.MetaData, nil
}

func (fsui *FileSystemUploadInspector) InspectUploadedFile(c context.Context, id string) (map[string]any, error) {
	filename := fsui.BaseDir + "/" + id
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, errors.Join(err, ErrNotFound)
	}
	uploadedFileInfo := map[string]any{
		"updated_at": fi.ModTime(),
		"size_bytes": fi.Size(),
	}
	return uploadedFileInfo, nil
}

func (aui *AzureUploadInspector) InspectInfoFile(c context.Context, id string) (map[string]any, error) {
	filename := id + ".info"
	infoBlobClient := aui.TusContainerClient.NewBlobClient(filename)

	// Download info file from blob client.
	downloadResponse, err := infoBlobClient.DownloadStream(c, nil)
	if err != nil {
		azErr, ok := err.(*azcore.ResponseError)
		if ok && azErr.StatusCode == http.StatusNotFound {
			return nil, errors.Join(err, ErrNotFound)
		}
		return nil, err
	}

	fileBytes, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		return nil, err
	}

	// Deserialize to hash map.
	jsonMap := &InfoFileData{}
	if err := json.Unmarshal(fileBytes, jsonMap); err != nil {
		return nil, err
	}

	return jsonMap.MetaData, nil
}

func (aui *AzureUploadInspector) InspectUploadedFile(c context.Context, id string) (map[string]any, error) {
	filename := id
	uploadBlobClient := aui.TusContainerClient.NewBlobClient(filename)
	propertiesResponse, err := uploadBlobClient.GetProperties(c, nil)
	if err != nil {
		return nil, errors.Join(err, ErrNotFound)
	}

	uploadedFileInfo := map[string]any{
		"updated_at": propertiesResponse.LastModified,
		"size_bytes": propertiesResponse.ContentLength,
	}

	return uploadedFileInfo, nil
}

type InfoResponse struct {
	Manifest map[string]any `json:"manifest"`
	FileInfo map[string]any `json:"file_info"`
}

type InfoHandler struct {
	inspecter UploadInspecter
}

func (ih *InfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("UploadID")

	fileInfo, err := ih.inspecter.InspectInfoFile(r.Context(), id)
	if err != nil {
		http.Error(rw, err.Error(), getStatusFromError(err))
		return
	}
	uploadedFileInfo, err := ih.inspecter.InspectUploadedFile(r.Context(), id)
	if err != nil {
		http.Error(rw, err.Error(), getStatusFromError(err))
		return
	}

	response := &InfoResponse{
		Manifest: fileInfo,
		FileInfo: uploadedFileInfo,
	}

	enc := json.NewEncoder(rw)
	enc.Encode(response)

}

func getStatusFromError(err error) int {
	if errors.Is(err, ErrNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func createInspector(appConfig *appconfig.AppConfig) (UploadInspecter, error) {
	if appConfig.TusAzStorageConfig != nil {
		// Create tus container client.
		containerClient, err := storeaz.NewContainerClient(*appConfig.TusAzStorageConfig)
		if err != nil {
			return nil, err
		}

		return NewAzureUploadInspector(containerClient, appConfig.TusdHandlerBasePath), nil
	}
	if appConfig.LocalFolderUploadsTus != "" {
		return NewFileSystemUploadInspector(appConfig.LocalFolderUploadsTus), nil
	}

	return nil, errors.New("Unable to create inspector given app configuration.")
}

func GetUploadInfoHandler(appConfig *appconfig.AppConfig) (http.Handler, error) {
	inspector, err := createInspector(appConfig)
	return &InfoHandler{
		inspector,
	}, err
}
