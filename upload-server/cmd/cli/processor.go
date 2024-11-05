package cli

import (
	"context"
	"fmt"

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

	if appConfig.SQSSubscriberConnection != nil {
		arn := appConfig.SQSSubscriberConnection.EventArn
		batchMax := appConfig.SQSSubscriberConnection.MaxMessages
		if batchMax == 0 {
			batchMax = event.MaxMessages
		}
		s, err := event.NewSQSSubscriber[T](ctx, arn, 1)
		if err != nil {
			return s, err
		}
		if err := s.Subscribe(ctx, arn); err != nil {
			return s, fmt.Errorf("arn: %s, %w", arn, err)
		}
		health.Register(s)
		return s, nil

	}

	if appConfig.SubscriberConnection != nil {
		sub, err := event.NewAzureSubscriber[T](ctx, appConfig.SubscriberConnection.ConnectionString, appConfig.SubscriberConnection.Topic, appConfig.SubscriberConnection.Subscription)
		if err != nil {
			return nil, err
		}

		health.Register(sub)
		return sub, nil
	}

	return sub, nil
}
