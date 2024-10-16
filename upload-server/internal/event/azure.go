package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"nhooyr.io/websocket"
)

func NewAMQPServiceBusClient(connString string) (*azservicebus.Client, error) {
	newWebSocketConnFn := func(ctx context.Context, args azservicebus.NewWebSocketConnArgs) (net.Conn, error) {
		opts := &websocket.DialOptions{Subprotocols: []string{"amqp"}}
		wssConn, _, err := websocket.Dial(ctx, args.Host, opts)
		if err != nil {
			return nil, err
		}

		return websocket.NetConn(ctx, wssConn, websocket.MessageBinary), nil
	}
	return azservicebus.NewClientFromConnectionString(connString, &azservicebus.ClientOptions{
		NewWebSocketConn: newWebSocketConnFn, // Setting this option so messages are sent to port 443.
	})
}

func NewAzurePublisher[T Identifiable](ctx context.Context, pubConn appconfig.AzureQueueConfig) (*AzurePublisher[T], error) {
	client, err := NewAMQPServiceBusClient(pubConn.ConnectionString)
	if err != nil {
		slog.Error("failed to connect to event service bus", "error", err)
		return nil, err
	}
	queueOrTopic := pubConn.Queue
	if queueOrTopic == "" {
		queueOrTopic = pubConn.Topic
	}
	sender, err := client.NewSender(queueOrTopic, nil)
	if err != nil {
		slog.Error("failed to configure event publisher", "error", err)
		return nil, err
	}
	adminClient, err := admin.NewClientFromConnectionString(pubConn.ConnectionString, nil)
	if err != nil {
		slog.Error("failed to connect to service bus admin client", "error", err)
		return nil, err
	}

	return &AzurePublisher[T]{
		Context:     ctx,
		Sender:      sender,
		Config:      pubConn,
		AdminClient: adminClient,
	}, nil
}

type AzurePublisher[T Identifiable] struct {
	Context     context.Context
	Sender      *azservicebus.Sender
	Config      appconfig.AzureQueueConfig
	AdminClient *admin.Client
}

func (ap *AzurePublisher[T]) Publish(ctx context.Context, event T) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ap.Sender.SendMessage(ctx, &azservicebus.Message{
		Body: b,
	}, nil)
}

func (ap *AzurePublisher[T]) Close() error {
	return ap.Sender.Close(ap.Context)
}

func (ap *AzurePublisher[T]) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE

	if ap.Config.Queue != "" {
		rsp.Service = fmt.Sprintf("Event Publishing %s", ap.Config.Queue)
		queueResp, err := ap.AdminClient.GetQueue(ctx, ap.Config.Queue, nil)
		if err != nil {
			return rsp.BuildErrorResponse(err)
		}
		if queueResp == nil {
			return rsp.BuildErrorResponse(fmt.Errorf("nil queue response"))
		}
		if *queueResp.Status != admin.EntityStatusActive {
			return rsp.BuildErrorResponse(fmt.Errorf("service bus queue %s status: %s", ap.Config.Queue, *queueResp.Status))
		}
	}

	if ap.Config.Topic != "" {
		rsp.Service = fmt.Sprintf("Event Publishing %s", ap.Config.Topic)
		topicResp, err := ap.AdminClient.GetTopic(ctx, ap.Config.Topic, nil)
		if err != nil {
			return rsp.BuildErrorResponse(err)
		}
		if *topicResp.Status != admin.EntityStatusActive {
			return rsp.BuildErrorResponse(fmt.Errorf("service bus topic %s status: %s", ap.Config.Topic, *topicResp.Status))
		}
	}

	return rsp
}

func NewAzureSubscriber[T Identifiable](ctx context.Context, subConn appconfig.AzureQueueConfig) (*AzureSubscriber[T], error) {
	client, err := NewAMQPServiceBusClient(subConn.ConnectionString)
	if err != nil {
		slog.Error("failed to connect to event service bus", "error", err)
		return nil, err
	}
	receiver, err := client.NewReceiverForSubscription(subConn.Topic, subConn.Subscription, nil)
	if err != nil {
		slog.Error("failed to configure event subscriber", "error", err)
		return nil, err
	}
	adminClient, err := admin.NewClientFromConnectionString(subConn.ConnectionString, nil)
	if err != nil {
		slog.Error("failed to connect to service bus admin client", "error", err)
		return nil, err
	}
	return &AzureSubscriber[T]{
		Context:     ctx,
		Receiver:    receiver,
		Config:      subConn,
		AdminClient: adminClient,
	}, nil
}

type AzureSubscriber[T Identifiable] struct {
	Context     context.Context
	Receiver    *azservicebus.Receiver
	Config      appconfig.AzureQueueConfig
	AdminClient *admin.Client
}

func (as *AzureSubscriber[T]) GetBatch(ctx context.Context, max int) ([]T, error) {
	msgs, err := as.Receiver.ReceiveMessages(ctx, max, nil)
	if err != nil {
		return nil, err
	}

	var batch []T
	for _, m := range msgs {
		slog.Info("received event", "event", m.Body)

		var e T
		e, err := NewEventFromServiceBusMessage[T](m)
		if err != nil {
			return nil, err
		}
		batch = append(batch, e)
	}

	return batch, nil
}

func (as *AzureSubscriber[T]) HandleSuccess(ctx context.Context, e T) error {
	if e.OrigMessage() == nil {
		return fmt.Errorf("malformed event %+v", e)
	}
	err := as.Receiver.CompleteMessage(ctx, e.OrigMessage(), nil)
	if err != nil {
		slog.Error("failed to ack event", "error", err)
		return err
	}
	slog.Info("successfully handled event", "event ID", e.Identifier(), "event type", e.Type())
	return nil
}

func (as *AzureSubscriber[T]) HandleError(ctx context.Context, e T, handlerError error) error {
	slog.Error("failed to handle event", "event ID", e.Identifier(), "event type", e.Type(), "error", handlerError.Error())
	return as.Receiver.DeadLetterMessage(ctx, e.OrigMessage(), nil)
}

func (as *AzureSubscriber[T]) Close() error {
	return as.Receiver.Close(as.Context)
}

func (as *AzureSubscriber[T]) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = fmt.Sprintf("%s Event Subscriber", as.Config.Subscription)
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE

	subResp, err := as.AdminClient.GetSubscription(ctx, as.Config.Topic, as.Config.Subscription, nil)
	if err != nil {
		return rsp.BuildErrorResponse(err)
	}

	if *subResp.Status != admin.EntityStatusActive {
		return rsp.BuildErrorResponse(fmt.Errorf("service bus subscription %s status: %s", as.Config.Subscription, *subResp.Status))
	}

	return rsp
}
