package event

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
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
	client, err := azservicebus.NewClientFromConnectionString(connString, &azservicebus.ClientOptions{
		NewWebSocketConn: newWebSocketConnFn, // Setting this option so messages are sent to port 443.
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
