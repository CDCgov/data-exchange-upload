package event

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
)

type Publishers[T Identifiable] []Publisher[T]

func (p Publishers[T]) Publish(ctx context.Context, e T) error {
	var errs error
	for _, publisher := range p {
		for range MaxRetries {
			if err := publisher.Publish(ctx, e); err != nil {
				slog.Error("Failed to publish", "event", e, "publisher", publisher, "err", err)
				errs = fmt.Errorf("Failed to publish event %s %w", e.Identifier(), err)
				continue
			}
			break
		}
	}
	return errs
}

func (p Publishers[T]) Close() {
	for _, publisher := range p {
		c, ok := publisher.(io.Closer)
		if ok {
			c.Close()
		}
	}
}

func InitFileReadyPublisher(ctx context.Context, appConfig appconfig.AppConfig) error {
	p, err := NewEventPublisher[*FileReady](ctx, appConfig)
	FileReadyPublisher = p
	return err
}

type Publisher[T Identifiable] interface {
	Publish(ctx context.Context, event T) error
}

func NewEventPublisher[T Identifiable](ctx context.Context, appConfig appconfig.AppConfig) (Publishers[T], error) {
	p := Publishers[T]{}
	c, err := GetChannel[T]()
	if err != nil {
		return nil, err
	}

	if appConfig.PublisherConnection != nil {
		ap, err := NewAzurePublisher[T](ctx, *appConfig.PublisherConnection)
		if err != nil {
			return p, err
		}
		health.Register(ap)
		p = append(p, ap)
		return p, err
	}

	p = append(p, &MemoryPublisher[T]{
		Chan: c,
	}, &FilePublisher[T]{
		Dir: appConfig.LocalEventsFolder,
	})

	return p, nil
}
