package cli

import (
	"context"
	"net"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	azureloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/azure"
	fileloader "github.com/cdcgov/data-exchange-upload/upload-server/internal/loaders/file"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
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

	cache := &metadata.ConfigCache{
		Loader: &fileloader.FileConfigLoader{
			FileSystem: os.DirFS(appConfig.UploadConfigPath),
		},
	}

	manifestValidator := metadata.SenderManifestVerification{
		Configs: cache,
		Reporter: &filereporters.FileReporter{
			Dir: appConfig.LocalReportsFolder,
		},
	}

	postReceiveHook := metadata.HookEventHandler{
		Reporter: &filereporters.FileReporter{
			Dir: appConfig.LocalReportsFolder,
		},
	}

	postFinishHook := metadata.HookEventHandler{
		Reporter: &filereporters.FileReporter{
			Dir: appConfig.LocalReportsFolder,
		},
	}

	postCreateHook := metadata.HookEventHandler{
		Reporter: &filereporters.FileReporter{
			Dir: appConfig.LocalReportsFolder,
		},
	}

	if appConfig.AzureConnection != nil {
		client, err := storeaz.NewBlobClient(*appConfig.AzureConnection)
		if err != nil {
			return nil, err
		}
		cache.Loader = &azureloader.AzureConfigLoader{
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

			// TODO Can these reporters all be pointers to a single instance?
			 manifestValidator.Reporter = &azurereporters.ServiceBusReporter{
				Client:    sbclient,
				QueueName: appConfig.ReportQueueName,
			}
			postReceiveHook.Reporter = &azurereporters.ServiceBusReporter{
				Client:    sbclient,
				QueueName: appConfig.ReportQueueName,
			}
			postFinishHook.Reporter = &azurereporters.ServiceBusReporter{
				Client:    sbclient,
				QueueName: appConfig.ReportQueueName,
			}
			postCreateHook.Reporter = &azurereporters.ServiceBusReporter{
				Client:    sbclient,
				QueueName: appConfig.ReportQueueName,
			} 			

		}
	}

	handler.Register(tusHooks.HookPreCreate, metadata.WithUploadID, metadata.WithTimestamp, manifestValidator.Verify)
	handler.Register(tusHooks.HookPostReceive, postReceiveHook.PostReceive)
	handler.Register(tusHooks.HookPostFinish, postFinishHook.PostFinish)
	handler.Register(tusHooks.HookPostCreate, postCreateHook.PostCreate)
	// TODO: -> handler.Register(tusHooks.HookPostFinish, copier.Merge, copier.Route)

	return handler, nil
}
