package cli

import (
	"context"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
)

func InitReporters(ctx context.Context, appConfig appconfig.AppConfig) error {
	reports.Register(&event.FilePublisher[*reports.Report]{
		Dir: appConfig.LocalReportsFolder,
	})

	if appConfig.SNSReporterConnection != nil {
		r, err := event.NewSNSPublisher[*reports.Report](ctx, appConfig.SNSReporterConnection.EventArn)
		if err != nil {
			return err
		}
		reports.Register(r)
		health.Register(r)
	}

	if appConfig.ReporterConnection != nil && appConfig.ReporterConnection.ConnectionString != "" {
		r, err := event.NewAzurePublisher[*reports.Report](ctx, *appConfig.ReporterConnection)
		if err != nil {
			return err
		}

		reports.Register(r)
		health.Register(r)
	}

	return nil
}
