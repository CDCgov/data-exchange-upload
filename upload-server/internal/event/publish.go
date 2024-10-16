package event

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
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

	if os.Getenv("SNS_EVENT_TOPIC_ARN") != "" {
		snsPub, err := NewSNSPublisher[T](ctx, sns.Options{
			Region:       os.Getenv("SNS_AWS_REGION"),
			BaseEndpoint: aws.String(os.Getenv("SNS_AWS_ENDPOINT_URL")),
			Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     os.Getenv("SNS_AWS_ACCESS_KEY_ID"),
					SecretAccessKey: os.Getenv("SNS_AWS_SECRET_ACCESS_KEY"),
				}, nil
			}),
		}, os.Getenv("SNS_EVENT_TOPIC_ARN"))
		if err != nil {
			return p, err
		}
		p = append(p, snsPub)
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

	if len(p) < 1 {
		c, err := GetChannel[T]()
		if err != nil {
			return nil, err
		}
		p = append(p, &MemoryPublisher[T]{
			Chan: c,
		}, &FilePublisher[T]{
			Dir: appConfig.LocalEventsFolder,
		})
	}

	return p, nil
}
