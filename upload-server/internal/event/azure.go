package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
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

func NewAzurePublisher[T Identifiable](ctx context.Context, connectionString, topic string) (*AzurePublisher[T], error) {
	client, err := NewAMQPServiceBusClient(connectionString)
	if err != nil {
		slog.Error("failed to connect to event service bus", "error", err)
		return nil, err
	}
	sender, err := client.NewSender(topic, nil)
	if err != nil {
		slog.Error("failed to configure event publisher", "error", err)
		return nil, err
	}
	adminClient, err := admin.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		slog.Error("failed to connect to service bus admin client", "error", err)
		return nil, err
	}

	return &AzurePublisher[T]{
		Context:     ctx,
		Sender:      sender,
		AdminClient: adminClient,
		Topic:       topic,
	}, nil
}

type AzurePublisher[T Identifiable] struct {
	Context     context.Context
	Sender      *azservicebus.Sender
	AdminClient *admin.Client
	Topic       string
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

	rsp.Service = fmt.Sprintf("Event Publishing %s", ap.Topic)
	topicResp, err := ap.AdminClient.GetTopic(ctx, ap.Topic, nil)
	if err != nil {
		return rsp.BuildErrorResponse(err)
	}
	if *topicResp.Status != admin.EntityStatusActive {
		return rsp.BuildErrorResponse(fmt.Errorf("service bus topic %s status: %s", ap.Topic, *topicResp.Status))
	}

	return rsp
}

func NewAzureSubscriber[T Identifiable](ctx context.Context, connectionString, topic, subscription string, maxMessages int) (*AzureSubscriber[T], error) {
	client, err := NewAMQPServiceBusClient(connectionString)
	if err != nil {
		slog.Error("failed to connect to event service bus", "error", err)
		return nil, err
	}
	receiver, err := client.NewReceiverForSubscription(topic, subscription, nil)
	if err != nil {
		slog.Error("failed to configure event subscriber", "error", err)
		return nil, err
	}
	adminClient, err := admin.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		slog.Error("failed to connect to service bus admin client", "error", err)
		return nil, err
	}

	if maxMessages == 0 {
		maxMessages = MaxMessages
	}

	return &AzureSubscriber[T]{
		Context:      ctx,
		Receiver:     receiver,
		AdminClient:  adminClient,
		Topic:        topic,
		Subscription: subscription,
		Max:          maxMessages,
	}, nil
}

func NewEventFromServiceBusMessage[T Identifiable](m *azservicebus.ReceivedMessage) (T, error) {
	var e T
	err := json.Unmarshal(m.Body, &e)
	if err != nil {
		return e, err
	}

	e.SetIdentifier(m.MessageID)

	return e, nil
}

type AzureSubscriber[T Identifiable] struct {
	Context      context.Context
	Receiver     *azservicebus.Receiver
	AdminClient  *admin.Client
	Subscription string
	Topic        string
	Max          int
}

func (as *AzureSubscriber[T]) Listen(ctx context.Context, process func(context.Context, T) error) error {
	defer as.Close()
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msgs, err := as.Receiver.ReceiveMessages(ctx, as.Max, nil)
			if err != nil {
				return err
			}

			for _, m := range msgs {
				slog.Info("received event", "event", m.Body)

				var e T
				e, err := NewEventFromServiceBusMessage[T](m)
				if err != nil {
					slog.Error("failed to get event from service bus", "message", m, "error", err.Error())
					if err := as.Receiver.DeadLetterMessage(ctx, m, nil); err != nil {
						slog.Error("failed to dead letter message", "message", m, "error", err.Error())
					}
				}
				if err := process(ctx, e); err != nil {
					if err := as.Receiver.DeadLetterMessage(ctx, m, nil); err != nil {
						slog.Error("failed to dead letter message", "message", m, "error", err.Error())
					}
					continue
				}
				if err := as.Receiver.CompleteMessage(ctx, m, nil); err != nil {
					slog.Error("failed to ack event", "error", err)
				}
			}
		}
	}
}

func (as *AzureSubscriber[T]) Close() error {
	return as.Receiver.Close(as.Context)
}

func (as *AzureSubscriber[T]) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = fmt.Sprintf("%s Event Subscriber", as.Subscription)
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE

	subResp, err := as.AdminClient.GetSubscription(ctx, as.Topic, as.Subscription, nil)
	if err != nil {
		return rsp.BuildErrorResponse(err)
	}

	if *subResp.Status != admin.EntityStatusActive {
		return rsp.BuildErrorResponse(fmt.Errorf("service bus subscription %s status: %s", as.Subscription, *subResp.Status))
	}

	return rsp
}

func (as *AzureSubscriber[T]) Length(ctx context.Context) (float64, error) {
	var l float64
	resp, err := as.AdminClient.GetSubscriptionRuntimeProperties(ctx, as.Topic, as.Subscription, nil)
	if err != nil {
		return l, nil
	}
	l = float64(resp.ActiveMessageCount)
	return l, nil
}

func (as *AzureSubscriber[T]) URL() string {
	return as.Subscription
}
