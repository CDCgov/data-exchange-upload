package cli

import (
	"context"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
)

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
