package cli

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"net"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	azureloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/azure"
	fileloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/file"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	azurereporters "github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters/azure"
	filereporters "github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters/file"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/storeaz"
	prebuilthooks "github.com/cdcgov/data-exchange-upload/upload-server/pkg/hooks"
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

func PrebuiltHooks(appConfig appconfig.AppConfig) (tusHooks.HookHandler, error) {
	handler := &prebuilthooks.PrebuiltHook{}

	metadata.Cache = &metadata.ConfigCache{
		Loader: &fileloader.FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	manifestValidator := metadata.SenderManifestVerification{
		Configs: metadata.Cache,
		Reporter: &filereporters.FileReporter{
			Dir: appConfig.LocalReportsFolder,
		},
	}

	hookHandler := metadata.HookEventHandler{
		Reporter: &filereporters.FileReporter{
			Dir: appConfig.LocalReportsFolder,
		},
	}

	postprocessing.RegisterTarget("dex", &postprocessing.FileDeliverer{
		ToPath: appConfig.LocalDEXFolder,
		From:   os.DirFS(appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix),
	})

	postprocessing.RegisterTarget("edav", &postprocessing.FileDeliverer{
		ToPath: appConfig.LocalEDAVFolder,
		From:   os.DirFS(appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix),
	})

	postprocessing.RegisterTarget("routing", &postprocessing.FileDeliverer{
		ToPath: appConfig.LocalROUTINGFolder,
		From:   os.DirFS(appConfig.LocalFolderUploadsTus + "/" + appConfig.TusUploadPrefix),
	})

	if appConfig.AzureConnection != nil {
		client, err := storeaz.NewBlobClient(*appConfig.AzureConnection)
		if err != nil {
			return nil, err
		}
		metadata.Cache.Loader = &azureloader.AzureConfigLoader{
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

			manifestValidator.Reporter = &azurereporters.ServiceBusReporter{
				Client:    sbclient,
				QueueName: appConfig.ReportQueueName,
			}
			hookHandler.Reporter = &azurereporters.ServiceBusReporter{
				Client:    sbclient,
				QueueName: appConfig.ReportQueueName,
			}

			// TODO use env vars for container names.
			tusContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, appConfig.AzureUploadContainer)
			if err != nil {
				return nil, err
			}
			dexCheckpointContainerClient, err := storeaz.NewContainerClient(*appConfig.AzureConnection, "dex-checkpoint")
			if err != nil {
				return nil, err
			}

			_, err = dexCheckpointContainerClient.GetProperties(context.TODO(), nil)
			if err != nil {
				var storageErr *azcore.ResponseError
				if errors.As(err, &storageErr) {
					if storageErr.StatusCode == http.StatusNotFound {
						logger.Info("creating dex-checkpoint container")
						_, err := dexCheckpointContainerClient.Create(context.TODO(), nil)
						if err != nil {
							logger.Error("failed to create dex checkpoint container")
							return nil, err
						}
					}
				}
			}

			postprocessing.RegisterTarget("dex", &postprocessing.AzureDeliverer{
				FromContainerClient: tusContainerClient,
				ToContainerClient:   dexCheckpointContainerClient,
			})
		}
	}

	handler.Register(tusHooks.HookPreCreate, hookHandler.WithUploadID, hookHandler.WithTimestamp, manifestValidator.Verify)
	handler.Register(tusHooks.HookPostReceive, hookHandler.PostReceive)
	handler.Register(tusHooks.HookPostCreate, hookHandler.PostCreate)
	// note that tus sends this to a potentially blocking channel.
	// however it immediately pulls from that channel in to a goroutine..so we're good
	handler.Register(tusHooks.HookPostFinish, hookHandler.PostFinish, manifestValidator.Hydrate, postprocessing.RouteAndDeliverHook)

	return handler, nil
}
