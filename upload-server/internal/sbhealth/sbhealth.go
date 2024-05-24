package sbhealth

import (
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

// Service Bus
type ServiceBusHealth struct {
	Client *admin.Client
	Queue  string
}

func New(appConfig appconfig.AppConfig) (*ServiceBusHealth, error) {
	// TODO set retry options.
	client, err := admin.NewClientFromConnectionString(appConfig.ServiceBusConnectionString, nil)
	if err != nil {
		return nil, err
	}

	return &ServiceBusHealth{
		Client: client,
		Queue:  appConfig.ReportQueueName,
	}, nil
}
