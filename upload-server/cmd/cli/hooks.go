package cli

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/handler"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/hooks/file"
)

func GetHookHandler(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	if Flags.FileHooksDir != "" {
		return &file.FileHook{
			Directory: Flags.FileHooksDir,
		}, nil
	}
	return PrebuiltHooks(appConfig)
}

func HookHandlerFunc(f func(handler.HookEvent) (handler.HTTPResponse, handler.FileInfoChanges, error)) func(handler.HookEvent) (tusHooks.HookResponse, error) {
	return func(e handler.HookEvent) (res tusHooks.HookResponse, err error) {
		resp, changes, err := f(e)
		res.HTTPResponse = resp
		res.ChangeFileInfo = changes
		return res, err
	}
}

type FileConfigLoader struct {
	FileSystem fs.FS
}

func (l *FileConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {

	file, err := l.FileSystem.Open(path)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(file)
}

type AzureConfigLoader struct {
	Client        *azblob.Client
	ContainerName string
}

func (l *AzureConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {
	downloadResponse, err := l.Client.DownloadStream(ctx, l.ContainerName, path, nil)
	if err != nil {
		var respErr *azcore.ResponseError
		if errors.As(err, &respErr) {
			if respErr.StatusCode == http.StatusNotFound {
				return nil, os.ErrNotExist
			}
		}
		return nil, err
	}

	return io.ReadAll(downloadResponse.Body)
}

func getRequiredMetadata(metadata map[string]interface{}) ([]interface{}, error) {
	metadataVersion := metadata["version"].(string)

	var requiredFields []string
		switch metadataVersion {
		case METADATA_VERSION_ONE:
			requiredFields = REQUIRED_VERSION_ONE_FIELDS
		case METADATA_VERSION_TWO:
			requiredFields = REQUIRED_VERSION_TWO_FIELDS
		default:
			return nil, fmt.Errorf("Unsupported metadata version: %s", metadataVersion)
	}

	var missingMetadataFields []interface{}
	for _, field := range requiredFields {
		if _, ok := metadata[field]; !ok {
			missingMetadataFields = append(missingMetadataFields, field)
		}
	}

	if len(missingMetadataFields) > 0 {
		return nil, fmt.Errorf("Missing one or more required metadata fields: %v", missingMetadataFields)
	}

	var values []interface{}
	for _, field := range requiredFields {
		values = append(values, metadata[field])
	}
	return values, nil
}

func getFilenameFromMetadata(metadata map[string]interface{}) string {
	filenameMetadataFields := []string{"filename", "original_filename", "meta_ext_filename"}

	var filename string
	for _, field := range filenameMetadataFields {
		if val, ok := metadata[field]; ok {
			filename = val.(string)
			break
		}
	}

	return filename
}

func postCreate(useCase, useCaseCategory string, metadata map[string]interface{}, tguid string) {
	logger.Printf("Creating trace for upload %s with use case %s and use case category %s\n", tguid, useCase, useCaseCategory)

	// send report message via service bus integration
	createMetadataVerificationReportJSON(metadata, tguid, useCase, useCaseCategory)
}

func createMetadataVerificationReportJSON(metadata map[string]interface{}, tguid, useCase, useCaseCategory string) {
	jsonPayload := map[string]interface{}{
		"schema_version": "0.0.1",
		"schema_name":    STAGE_NAME,
		"filename":       getFilenameFromMetadata(metadata),
		"timestamp":      time.Now().Format(time.RFC3339),
		"metadata":       metadata,
		"issues":         []interface{}{},
	}
}

func PrebuiltHooks(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	preCreateHook := metadata.SenderManifestVerification{
		Loader: &FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	if appConfig.DexAzUploadConfig != nil {
		client, err := storeaz.NewBlobClient(*appConfig.DexAzUploadConfig)
		if err != nil {
			return nil, err
		}
		preCreateHook.Loader = &AzureConfigLoader{
			Client:        client,
			ContainerName: appConfig.DexAzUploadConfig.AzContainerName,
		}
	}

	postCreateHook := HookHandlerFunc(func(e handler.HookEvent) (handler.HTTPResponse, handler.FileInfoChanges, error) {
		if e.Type == tusHooks.HookPreCreate {
			metadataJSON := map[string]interface{}{}
			err := json.Unmarshal(e.Upload.Metadata, &metadataJSON)
			if err != nil {
				return handler.HTTPResponse{}, nil, err
			}

			useCaseValues, err := getRequiredMetadata(metadataJSON)
			if err != nil {
				return handler.HTTPResponse{}, nil, err
			}

			postCreate(useCaseValues[0].(string), useCaseValues[1].(string), metadataJSON, e.Upload.ID)
		}

		return handler.HTTPResponse{}, nil, nil
	})

	handler.Register(tusHooks.HookPreCreate, preCreateHook.Verify)
	handler.Register(tusHooks.HookPostCreate, postCreateHook.Verify)
	return handler, nil
}
