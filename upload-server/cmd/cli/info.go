package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/azureinspector"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/fileinspector"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
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

func (ih *InfoHandler) Hook(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	id := event.Upload.ID
	ctx := event.Context

	var fileInfo map[string]any
	for i := 0; i < 5; i++ {
		var err error
		fileInfo, err = ih.inspecter.InspectInfoFile(ctx, id)
		if err != nil {
			logger.Info("Failed to validate manifest file after upload", "id", id, "error", err, "manifest", fileInfo)
		}
		if fileInfo != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}
	var uploadedFileInfo map[string]any
	for i := 0; i < 5; i++ {
		var err error
		uploadedFileInfo, err = ih.inspecter.InspectUploadedFile(ctx, id)
		if err != nil {
			logger.Info("Failed to validate upload file", "id", id, "error", err, "manifest", fileInfo, "upload", uploadedFileInfo)
		}
		if uploadedFileInfo != nil {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if fileInfo == nil {
		logger.Error("Failed to validate manifest file after upload", "id", id)
		return resp, errors.New("failed to find manifest")
	}

	if uploadedFileInfo == nil {
		logger.Error("Failed to validate upload file", "id", id, "manifest", fileInfo, "upload", uploadedFileInfo)
		return resp, errors.New("failed to find the uploaded file")
	}

	logger.Info("Upload validated", "id", id, "manifest", fileInfo, "upload_info", uploadedFileInfo)
	return resp, nil
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

func GetUploadInfoHandler(appConfig *appconfig.AppConfig) (*InfoHandler, error) {
	inspector, err := createInspector(appConfig)
	return &InfoHandler{
		inspector,
	}, err
}
