package cli

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
)

func InitReporters(ctx context.Context, appConfig appconfig.AppConfig) error {
	reports.DefaultReporter = &event.MemoryPublisher[*reports.Report]{
		Dir: appConfig.LocalReportsFolder,
	}

	if appConfig.ReporterConnection != nil && appConfig.ReporterConnection.ConnectionString != "" {
		r, err := event.NewAzurePublisher[*reports.Report](ctx, *appConfig.ReporterConnection, "Report")
		if err != nil {
			return err
		}

		reports.DefaultReporter = r
		health.Register(r)
	}

	return nil
}
