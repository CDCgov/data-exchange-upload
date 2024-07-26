package cli

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"sync"
)

func NewEventSubscriber(ctx context.Context, appConfig appconfig.AppConfig) event.Subscribable {
	var sub event.Subscribable
	sub = &event.MemorySubscriber{}

	if appConfig.SubscriberConnection != nil {
		client, err := event.NewAMQPServiceBusClient(appConfig.SubscriberConnection.ConnectionString)
		if err != nil {
			logger.Error("failed to connect to event service bus", "error", err)
		}
		receiver, err := client.NewReceiverForSubscription(appConfig.SubscriberConnection.Topic, appConfig.SubscriberConnection.Subscription, nil)
		if err != nil {
			logger.Error("failed to configure event subscriber", "error", err)
		}
		adminClient, err := admin.NewClientFromConnectionString(appConfig.PublisherConnection.ConnectionString, nil)
		if err != nil {
			logger.Error("failed to connect to service bus admin client", "error", err)
		}
		sub = &event.AzureSubscriber{
			Context:     ctx,
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
