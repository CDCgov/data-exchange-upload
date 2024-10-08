package event

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"net"
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
		logger.Error("failed to connect to event service bus", "error", err)
		return nil, err
	}
	queueOrTopic := pubConn.Queue
	if queueOrTopic == "" {
		queueOrTopic = pubConn.Topic
	}
	sender, err := client.NewSender(queueOrTopic, nil)
	if err != nil {
		logger.Error("failed to configure event publisher", "error", err)
		return nil, err
	}
	adminClient, err := admin.NewClientFromConnectionString(pubConn.ConnectionString, nil)
	if err != nil {
		logger.Error("failed to connect to service bus admin client", "error", err)
		return nil, err
	}

	return &AzurePublisher[T]{
		Context:     ctx,
		Sender:      sender,
		Config:      pubConn,
		AdminClient: adminClient,
	}, nil
}

func NewAzureSubscriber[T Identifiable](ctx context.Context, subConn appconfig.AzureQueueConfig) (*AzureSubscriber[T], error) {
	client, err := NewAMQPServiceBusClient(subConn.ConnectionString)
	if err != nil {
		logger.Error("failed to connect to event service bus", "error", err)
		return nil, err
	}
	receiver, err := client.NewReceiverForSubscription(subConn.Topic, subConn.Subscription, nil)
	if err != nil {
		logger.Error("failed to configure event subscriber", "error", err)
		return nil, err
	}
	adminClient, err := admin.NewClientFromConnectionString(subConn.ConnectionString, nil)
	if err != nil {
		logger.Error("failed to connect to service bus admin client", "error", err)
		return nil, err
	}
	return &AzureSubscriber[T]{
		Context:     ctx,
		Receiver:    receiver,
		Config:      subConn,
		AdminClient: adminClient,
	}, nil
}
