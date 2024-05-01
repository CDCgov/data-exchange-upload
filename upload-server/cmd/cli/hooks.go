package cli

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	tusHooks "github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/hooks/file"
	"nhooyr.io/websocket"
)

func GetHookHandler(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	if Flags.FileHooksDir != "" {
		return &file.FileHook{
			Directory: Flags.FileHooksDir,
		}, nil
	}
	return PrebuiltHooks(appConfig)
}

type FileConfigLoader struct {
	FileSystem fs.FS
}

func (l *FileConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {

	file, err := l.FileSystem.Open(path)
	if err != nil {
		return nil, errors.Join(err, validation.ErrNotFound)
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
				return nil, errors.Join(err, validation.ErrNotFound)
			}
		}
		return nil, err
	}

	return io.ReadAll(downloadResponse.Body)
}

type HookEventHandler struct {
	//Loader   validation.ConfigLoader
	Reporter metadata.Reporter
}

//type PostReceiveContent struct {
//	UploadId  string `json:"upload_id"`
//	StageName string `json:"stage_name"`
//}

func (v *HookEventHandler) postReceive(tguid string, offset int64, size int64, manifest map[string]string) error {

	logger.Info("go version", "version", runtime.Version())
	logger.Info("metadata values", "manifest", manifest)

	filename := metadata.GetFilename(manifest)

	logger.Info("file info", "filename", filename)

	//metadata_version = manifest.version
	//, METADATA_VERSION_ONE)
	//    json_data = get_report_body(metadata, metadata_version, filename, tguid, offset, size)

	//    logger.info('filename = {0}, metadata_version = {1}'.format(filename, metadata_version))

	//    logger.info('post_receive_bin: {0}, offset = {1}'.format(datetime.datetime.now(), offset))

	//    json_string = json.dumps(json_data)

	//    logger.info('JSON MESSAGE: %s', json_string)

	//    await send_message(json_string)

	//except Exception as e:
	//    logger.error("POST RECEIVE HOOK - exiting post_receive with error: %s", str(e), exc_info=True)
	//    sys.exit(1)
	return nil
}

// TODO: Relocate in to maybe internal/hooks or internal/upload-status ?
func (v *HookEventHandler) PostReceive(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {

	logger.Info("------resp-------", "resp", resp)

	// Get values from event
	uploadId := event.Upload.ID
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	uploadMetadata := event.Upload.MetaData

	logger.Info(
		"[PostReceive]: event.Upload values",
		"uploadMetadata", uploadMetadata,
		"uploadId", uploadId,
		"uploadSize", uploadSize,
		"uploadOffset", uploadOffset,
	)

	if err := v.postReceive(uploadId, uploadOffset, uploadSize, uploadMetadata); err != nil {
		//logger.Error("postReceive errors and warnings", "errors", err)
		logger.Error("postReceive errors and warnings", "err", err)

		//		content.Issues = &ValidationError{err}

		//		if errors.Is(err, validation.ErrFailure) {
		//			resp.RejectUpload = true
		//			resp.HTTPResponse = resp.HTTPResponse.MergeWith(handler.HTTPResponse{
		//				StatusCode: http.StatusBadRequest,
		//				Body:       err.Error(),
		//			})
		//			return resp, nil
		//		}
		//		return resp, err

	}

	return resp, nil

}

func PrebuiltHooks(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	preCreateHook := metadata.SenderManifestVerification{
		Loader: &FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
		Reporter: &metadata.FileReporter{
			Dir: appConfig.LocalReportsFolder,
		},
	}

	//postReceiveHook := metadata.SenderManifestVerification{
	postReceiveHook := HookEventHandler{
		//Loader: &FileConfigLoader{
		//	FileSystem: os.DirFS(appConfig.UploadConfigPath),
		//},
		Reporter: &metadata.FileReporter{
			Dir: appConfig.LocalReportsFolder,
		},
	}

	if appConfig.AzureConnection != nil {
		client, err := storeaz.NewBlobClient(*appConfig.AzureConnection)
		if err != nil {
			return nil, err
		}
		preCreateHook.Loader = &AzureConfigLoader{
			Client:        client,
			ContainerName: appConfig.AzureManifestConfigContainer,
		}

		if appConfig.ServiceBusConnectionString != "" {
			// Standard boilerplate for a websocket handler.
			newWebSocketConnFn := func(ctx context.Context, args azservicebus.NewWebSocketConnArgs) (net.Conn, error) {
				opts := &websocket.DialOptions{Subprotocols: []string{"amqp"}}
				wssConn, _, err := websocket.Dial(ctx, args.Host, opts)
				if err != nil {
					return nil, err
				}

				return websocket.NetConn(ctx, wssConn, websocket.MessageBinary), nil
			}
			sbclient, err := azservicebus.NewClientFromConnectionString(appConfig.ServiceBusConnectionString, &azservicebus.ClientOptions{
				NewWebSocketConn: newWebSocketConnFn, // Setting this option so messages are sent to port 443.
			})
			if err != nil {
				return nil, err
			}

			preCreateHook.Reporter = &metadata.ServiceBusReporter{
				Client:    sbclient,
				QueueName: appConfig.ReportQueueName,
			}
			postReceiveHook.Reporter = &metadata.ServiceBusReporter{
				Client:    sbclient,
				QueueName: appConfig.ReportQueueName,
			}
		}
	}

	handler.Register(tusHooks.HookPreCreate, metadata.WithUploadID, metadata.WithTimestamp, preCreateHook.Verify)
	handler.Register(tusHooks.HookPostReceive, postReceiveHook.PostReceive)

	return handler, nil
}
