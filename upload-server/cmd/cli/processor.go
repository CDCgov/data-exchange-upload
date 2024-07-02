package cli

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/aznamespaces"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"sync"
)

func MakeEventListener(appConfig appconfig.AppConfig, c chan event.FileReadyEvent) postprocessing.EventProcessable {
	var listener postprocessing.EventProcessable
	listener = &postprocessing.MemoryEventListener{
		C: c,
	}

	if appConfig.QueueConnection != nil {
		cred := azcore.NewKeyCredential(appConfig.QueueConnection.AccessKey)
		client, err := aznamespaces.NewReceiverClientWithSharedKeyCredential(appConfig.QueueConnection.Endpoint, appConfig.QueueConnection.Topic, appConfig.QueueConnection.Subscription, cred, nil)
		if err != nil {
			logger.Error("failed to configure azure receiver", "error", err)
		}
		listener = &postprocessing.AzureEventListener{
			Client: *client,
		}
	}

	return listener
}

func StartEventListener(ctx context.Context, listener postprocessing.EventProcessable) {
	for {
		var wg sync.WaitGroup
		events, err := listener.GetEventBatch(ctx, 5)
		if err != nil {
			// TODO dead letter
			continue
		}
		select {
		case <-ctx.Done():
			return
		default:
			for _, e := range events {
				wg.Add(1)
				e := e // TODO is this necessary?
				go func() {
					defer wg.Done()
					//listener.Process(ctx, e)

					err := postprocessing.ProcessFileReadyEvent(ctx, e)
					if err != nil {
						listener.HandleError(ctx, e, err)
						return
					}
					err = listener.HandleSuccess(ctx, e)
					if err != nil {
						listener.HandleError(ctx, e, err)
						return
					}

				}()
			}
			wg.Wait()
		}
	}

	//var wg sync.WaitGroup
	//numWorkers := len(workers)
	//wg.Add(numWorkers)
	//
	//for i := 0; i < numWorkers; i++ {
	//	i := i // TODO is this needed?
	//	go func() {
	//		//postprocessing.Worker(ctx, c)
	//		workers[i].DoWork(ctx)
	//		wg.Done()
	//	}()
	//}
	//
	//wg.Wait()
}
