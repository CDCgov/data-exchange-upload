package event

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snsTypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

var ErrInvalidARN = errors.New("given arn is not an arn")
var tags = map[string]string{
	"auto-generated":    "true",
	"auto-generated-by": "phdo",
}

func snsTags() (ts []snsTypes.Tag) {
	for k, v := range tags {
		ts = append(ts, snsTypes.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return ts
}

// arn format arn:aws:sns:region:account-id:topicname
func NewSNSPublisher[T Identifiable](ctx context.Context, topicARN string) (*SNSPublisher[T], error) {
	if !arn.IsARN(topicARN) {
		return nil, ErrInvalidARN
	}
	p := &SNSPublisher[T]{
		TopicArn: topicARN,
	}
	client, err := p.Client(ctx)
	if err != nil {
		return p, err
	}
	if _, err := client.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: aws.String(topicARN),
	}); err != nil {
		parsedARN, err := arn.Parse(topicARN)
		if err != nil {
			return nil, err
		}
		rsp, e := client.CreateTopic(ctx, &sns.CreateTopicInput{
			Name: aws.String(parsedARN.Resource),
			Tags: snsTags(),
		})
		if e != nil {
			return p, errors.Join(err, e)
		}
		slog.Info("Created topic", "arn", *rsp.TopicArn)
		p.TopicArn = *rsp.TopicArn
	}
	return p, nil
}

type SNSPublisher[T Identifiable] struct {
	TopicArn string
}

func (s SNSPublisher[T]) Client(ctx context.Context) (*sns.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return sns.NewFromConfig(cfg), nil
}

func (s *SNSPublisher[T]) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = fmt.Sprintf("SNS - %s", s.TopicArn)
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	client, err := s.Client(ctx)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if _, err := client.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: aws.String(s.TopicArn),
	}); err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	return rsp
}

func (s SNSPublisher[T]) Publish(ctx context.Context, e T) error {
	c, err := s.Client(ctx)
	if err != nil {
		return err
	}

	var b bytes.Buffer
	jsonEncoder := json.NewEncoder(&b)
	if err := jsonEncoder.Encode(e); err != nil {
		return err
	}
	m := b.String()
	result, err := c.Publish(ctx, &sns.PublishInput{
		Message:  &m,
		TopicArn: &s.TopicArn,
	})
	slog.Info("SNS event publish response", "response", result, "event", e)
	return err
}

func NewSQSSubscriber[T Identifiable](ctx context.Context, queueArn string, batchMax int, retryMax int) (*SQSSubscriber[T], error) {
	if !arn.IsARN(queueArn) {
		return nil, ErrInvalidARN
	}
	qa, err := arn.Parse(queueArn)
	if err != nil {
		return nil, err
	}
	s := &SQSSubscriber[T]{
		Max: batchMax,
		ARN: queueArn,
	}

	s.QueueURL, err = s.queue(ctx, qa.Resource)
	if err != nil {
		s.QueueURL, err = s.createQueue(ctx, qa.Resource)
		if err != nil {
			return s, err
		}
		if err := s.setUpDeadletterQueue(ctx, qa.Resource, retryMax); err != nil {
			return s, err
		}
	}

	arn := GetQueueArn(&s.QueueURL)
	if *arn != s.ARN {
		return s, fmt.Errorf("failed to get or create the correct queue arn, given arn: %s arn: %s url: %s", s.ARN, *arn, s.QueueURL)
	}

	return s, nil
}

type SQSSubscriber[T Identifiable] struct {
	QueueURL string
	ARN      string
	Max      int
}

func (s *SQSSubscriber[T]) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = fmt.Sprintf("SQS - %s", s.ARN)
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	client, err := s.Client(ctx)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if _, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(s.QueueURL),
	}); err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	return rsp
}

func (s *SQSSubscriber[T]) queue(ctx context.Context, name string) (string, error) {
	client, err := s.Client(ctx)
	if err != nil {
		return "", err
	}
	rsp, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		return "", err
	}
	return *rsp.QueueUrl, nil
}

func (s *SQSSubscriber[T]) createQueue(ctx context.Context, name string) (string, error) {
	client, err := s.Client(ctx)
	if err != nil {
		return "", err
	}
	rsp, e := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(name),
		Tags:      tags,
	})
	if e != nil {
		return "", e
	}
	return *rsp.QueueUrl, nil
}

func (s *SQSSubscriber[T]) setUpDeadletterQueue(ctx context.Context, name string, retryMax int) error {
	client, err := s.Client(ctx)
	if err != nil {
		return err
	}

	rsp, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(fmt.Sprintf("dl-queue-%s", name)),
		Tags:      tags,
	})
	if err != nil {
		return err
	}
	dlQueueARN := GetQueueArn(rsp.QueueUrl)

	policy := map[string]string{
		"deadLetterTargetArn": *dlQueueARN,
		"maxReceiveCount":     fmt.Sprintf("%d", retryMax),
	}

	b, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	// snippet-start:[sqs.go.dead_letter_queue.set_attributes]
	if _, err := client.SetQueueAttributes(ctx, &sqs.SetQueueAttributesInput{
		QueueUrl: aws.String(s.QueueURL),
		Attributes: map[string]string{
			"RedrivePolicy": string(b),
		},
	}); err != nil {
		return err
	}
	return nil
}

func GetQueueArn(queueURL *string) *string {
	// snippet-start:[sqs.go.get_queue_url.arn]
	parts := strings.Split(*queueURL, "/")
	subParts := strings.Split(parts[2], ".")

	arn := "arn:aws:" + subParts[0] + ":" + subParts[1] + ":" + parts[3] + ":" + parts[4]
	// snippet-end:[sqs.go.get_queue_url.arn]

	return &arn
}

func (s *SQSSubscriber[T]) Subscribe(ctx context.Context, topicArn string) error {
	if !arn.IsARN(topicArn) {
		return ErrInvalidARN
	}
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := sns.NewFromConfig(cfg)
	rsp, err := client.ListSubscriptionsByTopic(ctx, &sns.ListSubscriptionsByTopicInput{
		TopicArn: aws.String(topicArn),
	})
	if err != nil {
		return err
	}
	for _, sub := range rsp.Subscriptions {
		if *sub.Endpoint == s.QueueURL {
			slog.Info("Found subscription", "sub", sub, "url", s.QueueURL)
			return nil
		}
	}
	if _, err := client.Subscribe(ctx, &sns.SubscribeInput{
		Protocol: aws.String("sqs"),
		TopicArn: aws.String(topicArn),
		Endpoint: aws.String(s.ARN),
		Attributes: map[string]string{
			"RawMessageDelivery": "true",
		},
	}); err != nil {
		return err
	}
	return nil
}

func (s *SQSSubscriber[T]) Client(ctx context.Context) (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return sqs.NewFromConfig(cfg), nil
}

func (s *SQSSubscriber[T]) Listen(ctx context.Context, process func(context.Context, T) error) error {
	slog.Info("Listening to sqs queue", "queue", s.QueueURL)
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
				e, err := s.decodeEvent(&message)
				if err != nil {
					slog.Error("failed to decode message", "err", err.Error(), "message", message)
					if err := s.requeueMessage(ctx, message.ReceiptHandle); err != nil {
						slog.Error("failed to requeue message", "message", message, "error", err.Error())
					}
					continue
				}
				done := s.keepAlive(ctx, message.ReceiptHandle)
				if err := process(ctx, e); err != nil {
					slog.Error("failed to process message", "message", message, "error", err.Error())
					done()
					if err := s.requeueMessage(ctx, message.ReceiptHandle); err != nil {
						slog.Error("failed to requeue message", "message", message, "error", err.Error())
					}
					continue
				}
				done()
				s.deleteMessage(ctx, message.ReceiptHandle)
			}
		}
	}
}

var DefaultMessageVisibility = 30

func (s *SQSSubscriber[T]) getBatch(ctx context.Context, max int) ([]types.Message, error) {
	slog.Debug("Getting batch of messages from sqs", "max", max)
	svc, err := s.Client(ctx)
	if err != nil {
		return nil, err
	}

	rsp, err := svc.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.QueueURL),
		MaxNumberOfMessages: int32(max),
		WaitTimeSeconds:     1,
		VisibilityTimeout:   int32(DefaultMessageVisibility),
	})
	slog.Debug("Got sqs receive message response", "rsp", rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Messages, nil
}

func (s *SQSSubscriber[T]) keepAlive(ctx context.Context, handle *string) func() {
	c, cancel := context.WithCancel(ctx)
	go func(c context.Context, handle *string) {
		t := time.Tick(time.Duration(DefaultMessageVisibility-5) * time.Second)
		for {
			select {
			case <-t:
				svc, err := s.Client(c)
				if err != nil {
					slog.Error("failed to create client to keep ownership of message", "error", err, "handle", *handle)
				}
				_, err = svc.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
					QueueUrl:          aws.String(s.QueueURL),
					ReceiptHandle:     handle,
					VisibilityTimeout: int32(DefaultMessageVisibility),
				})
				if err != nil {
					slog.Error("failed to keep ownership of message", "error", err, "handle", *handle)
				}
			case <-c.Done():
				return
			}
		}
	}(c, handle)
	return cancel
}

func (s *SQSSubscriber[T]) decodeEvent(message *types.Message) (T, error) {
	var e T
	id := *message.MessageId
	body := *message.Body
	if err := json.Unmarshal([]byte(body), &e); err != nil {
		return e, err
	}
	e.SetIdentifier(id)
	return e, nil
}

func (s *SQSSubscriber[T]) requeueMessage(ctx context.Context, handle *string) error {
	svc, err := s.Client(ctx)
	if err != nil {
		return err
	}
	_, err = svc.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(s.QueueURL),
		ReceiptHandle:     handle,
		VisibilityTimeout: 0,
	})
	return err
}

func (s *SQSSubscriber[T]) deleteMessage(ctx context.Context, handle *string) error {
	svc, err := s.Client(ctx)
	if err != nil {
		return err
	}
	_, err = svc.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.QueueURL),
		ReceiptHandle: handle,
	})
	return err
}

func (s *SQSSubscriber[T]) Close() error {
	return nil
}
