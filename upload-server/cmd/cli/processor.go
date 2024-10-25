package cli

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"go.opentelemetry.io/otel"
	otrace "go.opentelemetry.io/otel/trace"
)

type key int

var UploadID key

func NewEventSubscriber[T event.Identifiable](ctx context.Context, appConfig appconfig.AppConfig) (event.Subscribable[T], error) {
	var sub event.Subscribable[T]
	c, err := event.GetChannel[T]()
	if err != nil {
		return nil, err
	}
	sub = &event.MemorySubscriber[T]{
		Chan: c,
	}

	if appConfig.SubscriberConnection != nil {
		sub, err := event.NewAzureSubscriber[T](ctx, *appConfig.SubscriberConnection)
		if err != nil {
			return nil, err
		}

		health.Register(sub)
		return sub, nil
	}

	return sub, nil
}

func SubscribeToEvents[T event.Identifiable](ctx context.Context, sub event.Subscribable[T], process func(context.Context, T) error) {
	tracer := otel.Tracer("event-handling")
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
				go func(ctx context.Context, e T) {
					defer wg.Done()
					_, span := tracer.Start(ctx, fmt.Sprintf("Handling-%s", e.Identifier()))
					defer span.End()
					err := process(ctx, e)

					if err != nil {
						logger.Error("failed to process event", "event", e, "error", err)
						err = sub.HandleError(ctx, e, err)
						if err != nil {
							logger.Error("failed to handle event error", "event", e, "error", err)
						}
						return
					}
					err = sub.HandleSuccess(ctx, e)
					if err != nil {
						logger.Error("failed to acknowledge event", "event", e, "error", err)
						err = sub.HandleError(ctx, e, err)
						if err != nil {
							logger.Error("failed to handle event error", "event", e, "error", err)
						}
						return
					}

				}(context.WithValue(ctx, UploadID, otrace.TraceID(md5.Sum([]byte(e.Identifier())))), e)
			}
			wg.Wait()
		}
	}
}
