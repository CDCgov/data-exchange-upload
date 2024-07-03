package cli

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/aznamespaces"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"sync"
)

func MakeEventSubscriber(appConfig appconfig.AppConfig) event.Subscribable {
	var sub event.Subscribable
	sub = &event.MemorySubscriber{}

	if appConfig.SubscriberConnection != nil {
		cred := azcore.NewKeyCredential(appConfig.SubscriberConnection.AccessKey)
		client, err := aznamespaces.NewReceiverClientWithSharedKeyCredential(appConfig.PublisherConnection.Endpoint, appConfig.PublisherConnection.Topic, appConfig.PublisherConnection.Subscription, cred, nil)
		if err != nil {
			logger.Error("failed to configure azure receiver", "error", err)
		}
		sub = &event.AzureSubscriber{
			Client: client,
			Config: *appConfig.SubscriberConnection,
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
