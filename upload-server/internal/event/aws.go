package event

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SNSPublisher[T Identifiable] struct {
	Options  sns.Options
	TopicArn string
}

func (s SNSPublisher[T]) Client() *sns.Client {
	return sns.New(s.Options)
}

func (s SNSPublisher[T]) Publish(ctx context.Context, e T) error {
	c := s.Client()

	var b bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &b)
	jsonEncoder := json.NewEncoder(encoder)
	if err := jsonEncoder.Encode(e); err != nil {
		return err
	}
	encoder.Close()
	m := b.String()
	result, err := c.Publish(ctx, &sns.PublishInput{
		Message:  &m,
		TopicArn: &s.TopicArn,
	})
	slog.Info("SNS event publish response", "response", result, "event", e)
	return err
}
