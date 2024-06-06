package cli

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	azurereporters "github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters/azure"
	filereporters "github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters/file"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"net"
	"nhooyr.io/websocket"
)

func InitReporters(appConfig appconfig.AppConfig) error {
	reports.DefaultReporter = &filereporters.FileReporter{
		Dir: appConfig.LocalReportsFolder,
	}

	if appConfig.AzureConnection != nil && appConfig.ServiceBusConnectionString != "" {
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
			return err
		}

		reports.DefaultReporter = &azurereporters.ServiceBusReporter{
			Client:    sbclient,
			QueueName: appConfig.ReportQueueName,
		}
	}

	return nil
}
