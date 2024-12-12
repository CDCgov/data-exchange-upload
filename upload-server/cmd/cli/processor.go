package cli

import (
	"context"
	"fmt"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
)

func NewEventSubscriber[T event.Identifiable](ctx context.Context, appConfig appconfig.AppConfig, defaultBus event.Subscribable[T]) (event.Subscribable[T], error) {

	if appConfig.SQSSubscriberConnection != nil {
		sqsArn := appConfig.SQSSubscriberConnection.EventArn
		snsArn := appConfig.SNSPublisherConnection.EventArn

		batchMax := appConfig.SQSSubscriberConnection.MaxMessages
		if batchMax == 0 {
			batchMax = event.MaxMessages
		}
		s, err := event.NewSQSSubscriber[T](ctx, sqsArn, 1)
		if err != nil {
			return s, err
		}
		if err := s.Subscribe(ctx, snsArn); err != nil {
			return s, fmt.Errorf("arn: %s, %w", snsArn, err)
		}
		health.Register(s)
		return s, nil

	}

	if appConfig.SubscriberConnection != nil {
		sub, err := event.NewAzureSubscriber[T](ctx, appConfig.SubscriberConnection.ConnectionString, appConfig.SubscriberConnection.Topic, appConfig.SubscriberConnection.Subscription, appConfig.SubscriberConnection.MaxMessages)
		if err != nil {
			return nil, err
		}

		health.Register(sub)
		return sub, nil
	}

	return defaultBus, nil
}
