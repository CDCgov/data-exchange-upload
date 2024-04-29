package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
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

// Placeholder for Azure Service Bus interaction
//func sendServiceBusMessage(message string) error {
//	// ... Your Azure Service Bus code to send the message ...
//	// TODO: Maybe just message to log stream for now for local dev?
//	return nil
//}

// TODO: Relocate in to maybe internal/hooks or internal/upload-status ?
func postReceiveHook(event handler.HookEvent) (hooks.HookResponse, error) {
	resp := hooks.HookResponse{}

	// Get values from event
	uploadId := event.Upload.ID
	uploadSize := event.Upload.Size
	uploadOffset := event.Upload.Offset
	uploadMetadata := event.Upload.MetaData

	logger.Info(
		"[PostReceive]: event.Upload values",
		"uploadMetadata", uploadMetadata,
		"uploadId", uploadId,
		"uploadSize", uploadSize,
		"uploadOffset", uploadOffset,
	)

	// filePath := fmt.Sprintf("/tmp/testing1111.txt")
	filePath := fmt.Sprintf("/tmp/%s.txt", uploadId)
	firstUpdate := true
	var elapsedSeconds int64 = 0

	// Check if the file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.Info("[post-receive]: Latest offset file NOT found")
	} else {
		logger.Info("[post-receive]: Found latest offset file")

		// Read the latest offset
		latestOffsetBytes, err := os.ReadFile(filePath)
		if err != nil {
			logger.Error("[post-receive]: ERROR: Failed to read offset:", "err", err)
		} else {
			logger.Info("[post-receive]: GOOD: Calling postReceive()")
		}
		latestOffsetStr := string(latestOffsetBytes)

		// Get file modification time
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			logger.Error("[post-receive]: ERROR: Failed to stat offset file:", "err", err)
		}
		lastModifiedEpoch := fileInfo.ModTime().Unix()

		// Get current time
		nowEpoch := time.Now().Unix()

		// Calculate elapsed seconds
		elapsedSeconds = nowEpoch - lastModifiedEpoch

		logger.Info("[post-receive]: Upload information.",
			"uploadOffset", uploadOffset,
			"latestOffset", latestOffsetStr,
			"nowEpoch", nowEpoch,
			"lastModifiedEpoch", lastModifiedEpoch,
			"elapsedSec", elapsedSeconds)
	}

	// Conditional Update Logic
	if firstUpdate || elapsedSeconds >= 1 || uploadOffset == uploadSize {

		logger.Info("[post-receive]: Updating latest offset file and processing update.",
			"offset", uploadOffset)

		uploadOffsetStr := strconv.FormatInt(uploadOffset, 10)
		// Convert string to byte slice
		data := []byte(uploadOffsetStr)

		err := os.WriteFile(filePath, data, 0644)
		if err != nil {
			// Handle error, likely log it or return
			logger.Info("[post-receive]: Error updating offset file:", "err", err)
		} else {
			logger.Info("[post-receive]: Latest offset file updated, calling post-receive-bin.",
				"offset", uploadOffset)
		}

		// TODO: Replace post-receive-bin.py Python with Go HERE
		//     ./post-receive-bin --id $id --offset $offset --size $size --metadata "$metadata"
		// create processPostReceive function
		//		- get metadata from event.Upload.MetaData
		//		- get filename from metadata
		//		- create JSON message
		//		- send JSON message
		// 		- Load Service Bus connection details from environment or config
		// 			connectionString := os.Getenv("SERVICE_BUS_CONNECTION_STRING")
		// 			queueName := os.Getenv("QUEUE_NAME")

	} else {
		logger.Info("[post-receive]: Skipping update")
	}

	return resp, nil

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

	handler.Register(tusHooks.HookPreCreate, preCreateHook.Verify)
	handler.Register(tusHooks.HookPostReceive, postReceiveHook)
	return handler, nil
}
