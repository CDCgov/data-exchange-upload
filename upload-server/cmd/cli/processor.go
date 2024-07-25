package cli

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"net"
	"nhooyr.io/websocket"
	"sync"
)

func NewEventSubscriber(appConfig appconfig.AppConfig) event.Subscribable {
	var sub event.Subscribable
	sub = &event.MemorySubscriber{}

	if appConfig.SubscriberConnection != nil {
		newWebSocketConnFn := func(ctx context.Context, args azservicebus.NewWebSocketConnArgs) (net.Conn, error) {
			opts := &websocket.DialOptions{Subprotocols: []string{"amqp"}}
			wssConn, _, err := websocket.Dial(ctx, args.Host, opts)
			if err != nil {
				return nil, err
			}

			return websocket.NetConn(ctx, wssConn, websocket.MessageBinary), nil
		}
		client, err := azservicebus.NewClientFromConnectionString(appConfig.SubscriberConnection.ConnectionString, &azservicebus.ClientOptions{
			NewWebSocketConn: newWebSocketConnFn, // Setting this option so messages are sent to port 443.
		})
		if err != nil {
			logger.Error("failed to connect to event service bus", "error", err)
		}
		//cred := azcore.NewKeyCredential(appConfig.SubscriberConnection.AccessKey)
		//client, err := aznamespaces.NewReceiverClientWithSharedKeyCredential(appConfig.PublisherConnection.Endpoint, appConfig.PublisherConnection.Topic, appConfig.PublisherConnection.Subscription, cred, nil)
		receiver, err := client.NewReceiverForSubscription(appConfig.SubscriberConnection.Topic, appConfig.SubscriberConnection.Subscription, nil)
		if err != nil {
			logger.Error("failed to configure event subscriber", "error", err)
		}
		adminClient, err := admin.NewClientFromConnectionString(appConfig.PublisherConnection.ConnectionString, nil)
		if err != nil {
			logger.Error("failed to connect to service bus admin client", "error", err)
		}
		sub = &event.AzureSubscriber{
			//Client: client,
			EventType:   event.FileReadyEventType,
			Receiver:    receiver,
			Config:      *appConfig.SubscriberConnection,
			AdminClient: adminClient,
		}

		health.Register(sub)
	}

	return sub
}

func SubscribeToEvents(ctx context.Context, sub event.Subscribable) {
	for {
		var wg sync.WaitGroup
		events, err := sub.GetBatch(ctx, 5)
		if err != nil {
			logger.Error("failed to get event batch", "error", err)
			continue
		}
		select {
		case <-ctx.Done():
			return
		default:
			for _, e := range events {
				wg.Add(1)
				go func(e event.FileReady) {
					defer wg.Done()
					err := postprocessing.ProcessFileReadyEvent(ctx, e)
					if err != nil {
						logger.Error("failed to process event", "event", e, "error", err)
						sub.HandleError(ctx, e, err)
						return
					}
					err = sub.HandleSuccess(ctx, e)
					if err != nil {
						logger.Error("failed to acknowledge event", "event", e, "error", err)
						sub.HandleError(ctx, e, err)
						return
					}

				}(e)
			}
			wg.Wait()
		}
	}
}
