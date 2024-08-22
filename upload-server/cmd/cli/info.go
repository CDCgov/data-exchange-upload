package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/stores3"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/inspector"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/azureinspector"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/fileinspector"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
)

type UploadInspector interface {
	InspectInfoFile(c context.Context, id string) (map[string]any, error)
	InspectUploadedFile(c context.Context, id string) (map[string]any, error)
}

type InfoHandler struct {
	inspector UploadInspector
}

func (ih *InfoHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	id := r.PathValue("UploadID")

	fileInfo, err := ih.inspector.InspectInfoFile(r.Context(), id)
	if err != nil {
		http.Error(rw, "error getting file manifest", getStatusFromError(err))
		return
	}

	response := &info.InfoResponse{
		Manifest: fileInfo,
	}

	uploadedFileInfo, err := ih.inspector.InspectUploadedFile(r.Context(), id)
	if err != nil {
		// skip not found errors to handle deferred uploads.
		if !errors.Is(err, inspector.ErrUploadNotFound) {
			http.Error(rw, fmt.Sprintf("error getting file info.  Manifest: %#v", fileInfo), getStatusFromError(err))
			return
		}
	}

	response.FileInfo = uploadedFileInfo

	enc := json.NewEncoder(rw)
	enc.Encode(response)
}

func getStatusFromError(err error) int {
	if errors.Is(err, info.ErrNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

func createInspector(ctx context.Context, appConfig *appconfig.AppConfig) (UploadInspector, error) {
	if appConfig.AzureConnection != nil {
		// Create tus container client.
		containerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
		if err != nil {
			return nil, err
		}

		return azureinspector.NewAzureUploadInspector(containerClient, appConfig.TusUploadPrefix), nil
	}
	if appConfig.S3Connection != nil {
		s3Client, err := stores3.New(ctx, appConfig.S3Connection)
		if err != nil {
			return nil, err
		}

		return &inspector.S3UploadInspector{
			Client:     s3Client,
			BucketName: appConfig.S3Connection.BucketName,
			TusPrefix:  appConfig.TusUploadPrefix,
		}, nil
	}
	if appConfig.LocalFolderUploadsTus != "" {
		return fileinspector.NewFileSystemUploadInspector(appConfig.LocalFolderUploadsTus, appConfig.TusUploadPrefix), nil
	}

	return nil, errors.New("unable to create inspector given app configuration")
}

func GetUploadInfoHandler(ctx context.Context, appConfig *appconfig.AppConfig) (http.Handler, error) {
	i, err := createInspector(ctx, appConfig)
	return &InfoHandler{
		i,
	}, err
}
