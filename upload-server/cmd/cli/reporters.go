package cli

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
)

func InitReporters(ctx context.Context, appConfig appconfig.AppConfig) error {
	//reports.DefaultReporter = &filereporters.FileReporter{
	//	Dir: appConfig.LocalReportsFolder,
	//}
	reports.DefaultReporter = &event.MemoryPublisher[reports.Report]{
		Dir: appConfig.LocalReportsFolder,
	}

	if appConfig.ReporterConnection != nil && appConfig.ReporterConnection.ConnectionString != "" {
		//sbClient, err := event.NewAMQPServiceBusClient(appConfig.ReporterConnection.ConnectionString)
		//if err != nil {
		//	return err
		//}
		//sender, err := sbClient.NewSender(appConfig.ReporterConnection.Queue, nil)
		//if err != nil {
		//	logger.Error("failed to configure report publisher", "error", err)
		//	return err
		//}
		//adminClient, err := admin.NewClientFromConnectionString(appConfig.ReporterConnection.ConnectionString, nil)
		//if err != nil {
		//	logger.Error("failed to connect to service bus admin client", "error", err)
		//	return err
		//}
		//
		//r := &azurereporters.ServiceBusReporter{
		//	Context:     ctx,
		//	Sender:      sender,
		//	AdminClient: adminClient,
		//	QueueName:   appConfig.ReporterConnection.Queue,
		//}

		r, err := event.NewAzurePublisher[reports.Report](ctx, *appConfig.ReporterConnection, "Report")
		if err != nil {
			return err
		}

		reports.DefaultReporter = r
		health.Register(r)
	}

	return nil
}
