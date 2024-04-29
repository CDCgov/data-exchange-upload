package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/azureinspector"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/fileinspector"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
)

type UploadInspecter interface {
	InspectInfoFile(c context.Context, id string) (map[string]any, error)
	InspectUploadedFile(c context.Context, id string) (map[string]any, error)
}

type InfoHandler struct {
	inspecter UploadInspecter
}

func (ih *InfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("UploadID")

	fileInfo, err := ih.inspecter.InspectInfoFile(r.Context(), id)
	if err != nil {
		http.Error(rw, "error getting file manifest", getStatusFromError(err))
		return
	}
	uploadedFileInfo, err := ih.inspecter.InspectUploadedFile(r.Context(), id)
	if err != nil {
		http.Error(rw, fmt.Sprintf("error getting file info.  Manifest: %#v", fileInfo), getStatusFromError(err))
		return
	}

	response := &info.InfoResponse{
		Manifest: fileInfo,
		FileInfo: uploadedFileInfo,
	}

	enc := json.NewEncoder(rw)
	enc.Encode(response)
}

func getStatusFromError(err error) int {
	if errors.Is(err, info.ErrNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func createInspector(appConfig *appconfig.AppConfig) (UploadInspecter, error) {
	if appConfig.AzureConnection != nil {
		// Create tus container client.
		containerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
		if err != nil {
			return nil, err
		}

		return azureinspector.NewAzureUploadInspector(containerClient, appConfig.TusUploadPrefix), nil
	}
	if appConfig.LocalFolderUploadsTus != "" {
		return fileinspector.NewFileSystemUploadInspector(appConfig.LocalFolderUploadsTus, appConfig.TusUploadPrefix), nil
	}

	return nil, errors.New("unable to create inspector given app configuration")
}

func GetUploadInfoHandler(appConfig *appconfig.AppConfig) (http.Handler, error) {
	inspector, err := createInspector(appConfig)
	return &InfoHandler{
		inspector,
	}, err
}
