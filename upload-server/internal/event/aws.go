package event

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

type SQSSubscriber[T Identifiable] struct {
	Config   sqs.Options
	QueueURL string
	Max      int
}

func (s *SQSSubscriber[T]) Client() *sqs.Client {
	return sqs.New(s.Config)
}

func (s *SQSSubscriber[T]) Listen(ctx context.Context, process func(context.Context, T) error) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			messages, err := s.getBatch(ctx, s.Max)
			if err != nil {
				return err
			}
			for _, message := range messages {
				e, err := s.decodeEvent(message)
				if err != nil {
					//TODO maybe immediately requeue the whole batch?
					return err
				}
				if err := process(ctx, e); err != nil {
					s.requeueMessage(ctx, message.ReceiptHandle)
					//TODO log error
					continue
				}
				s.deleteMessage(ctx, message.ReceiptHandle)
			}
		}
	}
}

func (s *SQSSubscriber[T]) getBatch(ctx context.Context, max int) ([]types.Message, error) {
	svc := s.Client()

	rsp, err := svc.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.QueueURL),
		MaxNumberOfMessages: int32(max),
		//VisibilityTimeout:   timeout,
	})
	if err != nil {
		return nil, err
	}
	return rsp.Messages, nil
}

func (s *SQSSubscriber[T]) decodeEvent(message types.Message) (T, error) {
	id := *message.MessageId
	body := *message.Body
	b := bytes.NewBuffer([]byte(body))
	decoder := base64.NewDecoder(base64.StdEncoding, b)
	jsonDecoder := json.NewDecoder(decoder)
	var e T
	if err := jsonDecoder.Decode(&e); err != nil {
		return e, err
	}
	e.SetIdentifier(id)
	return e, nil
}

func (s *SQSSubscriber[T]) requeueMessage(ctx context.Context, handle *string) error {
	svc := s.Client()
	_, err := svc.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(s.QueueURL),
		ReceiptHandle:     handle,
		VisibilityTimeout: 0,
	})
	//TODO log resp
	return err
}

func (s *SQSSubscriber[T]) deleteMessage(ctx context.Context, handle *string) error {
	svc := s.Client()
	_, err := svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.QueueURL),
		ReceiptHandle: handle,
	})
	//TODO log resp
	return err
}
