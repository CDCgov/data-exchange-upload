package reporters

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters"
)

type ServiceBusReporter struct {
	Client    *azservicebus.Client
	QueueName string
}

func (sb *ServiceBusReporter) Publish(ctx context.Context, r reporters.Identifiable) error {
	if sb.Client == nil {
		return errors.New("misconfigured Service Bus Reporter, missing client")
	}
	sender, err := sb.Client.NewSender(sb.QueueName, nil)
	if err != nil {
		return err
	}
	defer sender.Close(ctx)

	b, err := json.Marshal(r)
	if err != nil {
		return err
	}

	m := &azservicebus.Message{
		Body: b,
	}

	return sender.SendMessage(ctx, m, nil)
}
