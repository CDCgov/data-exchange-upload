package cli

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
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

	if os.Getenv("SQS_EVENT_ARN") != "" {
		s, err := event.NewSQSSubscriber[T](ctx, os.Getenv("SQS_EVENT_ARN"), 1, sqs.Options{
			Region:       os.Getenv("SNS_AWS_REGION"),
			BaseEndpoint: aws.String(os.Getenv("SNS_AWS_ENDPOINT_URL")),
			Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     os.Getenv("SNS_AWS_ACCESS_KEY_ID"),
					SecretAccessKey: os.Getenv("SNS_AWS_SECRET_ACCESS_KEY"),
				}, nil
			}),
		})
		if err != nil {
			return s, err
		}
		if err := s.Subscribe(ctx, os.Getenv("SNS_EVENT_TOPIC_ARN"), sns.Options{
			Region:       os.Getenv("SNS_AWS_REGION"),
			BaseEndpoint: aws.String(os.Getenv("SNS_AWS_ENDPOINT_URL")),
			Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     os.Getenv("SNS_AWS_ACCESS_KEY_ID"),
					SecretAccessKey: os.Getenv("SNS_AWS_SECRET_ACCESS_KEY"),
				}, nil
			}),
		}); err != nil {
			return s, err
		}
		return s, nil

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
