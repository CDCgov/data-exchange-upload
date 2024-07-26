package cli

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"sync"
)

func NewEventSubscriber[T event.Identifiable](ctx context.Context, appConfig appconfig.AppConfig, memChan chan T) (event.Subscribable[T], error) {
	var sub event.Subscribable[T]
	sub = &event.MemorySubscriber[T]{
		Chan: memChan,
	}

	if appConfig.SubscriberConnection != nil {
		sub, err := event.NewAzureSubscriber[event.FileReady](ctx, *appConfig.SubscriberConnection, event.FileReadyEventType)
		if err != nil {
			return nil, err
		}

		health.Register(sub)
	}

	return sub, nil
}

func SubscribeToEvents[T event.Identifiable](ctx context.Context, sub event.Subscribable[T]) {
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
				go func(e T) {
					defer wg.Done()
					if e, ok := event.Identifiable(e).(event.FileReady); ok {
						err = postprocessing.ProcessFileReadyEvent(ctx, e)
					}

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
