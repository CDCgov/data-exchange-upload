package cli

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	azurereporters "github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters/azure"
	filereporters "github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters/file"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
)

func InitReporters(ctx context.Context, appConfig appconfig.AppConfig) error {
	reports.DefaultReporter = &filereporters.FileReporter{
		Dir: appConfig.LocalReportsFolder,
	}

	if appConfig.AzureConnection != nil && appConfig.ServiceBusConnectionString != "" {
		sbClient, err := event.NewAMQPServiceBusClient(appConfig.ServiceBusConnectionString)
		if err != nil {
			return err
		}
		sender, err := sbClient.NewSender(appConfig.ReportQueueName, nil)
		if err != nil {
			logger.Error("failed to configure report publisher", "error", err)
		}
		adminClient, err := admin.NewClientFromConnectionString(appConfig.PublisherConnection.ConnectionString, nil)
		if err != nil {
			logger.Error("failed to connect to service bus admin client", "error", err)
		}

		r := &azurereporters.ServiceBusReporter{
			Context:     ctx,
			Sender:      sender,
			AdminClient: adminClient,
			QueueName:   appConfig.ReportQueueName,
		}

		reports.DefaultReporter = r
		health.Register(r)
	}

	return nil
}
