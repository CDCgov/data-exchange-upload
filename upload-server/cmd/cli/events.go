package cli

import (
	"context"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
)

func NewEventPublisher[T event.Identifiable](ctx context.Context, appConfig appconfig.AppConfig, defaultBus event.Publisher[T]) (event.Publishers[T], error) {
	p := event.Publishers[T]{}

	if appConfig.SNSPublisherConnection != nil {
		arn := appConfig.SNSPublisherConnection.EventArn
		snsPub, err := event.NewSNSPublisher[T](ctx, arn)
		if err != nil {
			return p, err
		}
		p = append(p, snsPub)
	}

	if appConfig.PublisherConnection != nil {
		topic := appConfig.PublisherConnection.Queue
		if topic == "" {
			topic = appConfig.PublisherConnection.Topic
		}
		ap, err := event.NewAzurePublisher[T](ctx, appConfig.PublisherConnection.ConnectionString, topic)
		if err != nil {
			return p, err
		}
		health.Register(ap)
		p = append(p, ap)
	}

	if len(p) < 1 {
		p = append(p,
			defaultBus,
			&event.FilePublisher[T]{
				Dir: appConfig.LocalEventsFolder,
			})
	}

	return p, nil
}
