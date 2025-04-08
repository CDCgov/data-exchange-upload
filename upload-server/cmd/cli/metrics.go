package cli

// tusd metrics loading: https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
// note: tusd.Handler exposes metrics by cli flag and defaults true

import (
	"context"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tus/tusd/v2/pkg/hooks"
) // .import

func setupMetrics(ctx context.Context, pollInterval time.Duration, m ...prometheus.Collector) {
	metrics.RegisterMetrics(hooks.MetricsHookErrorsTotal, hooks.MetricsHookInvocationsTotal, metrics.HttpReqs, metrics.OpenConnections, metrics.ActiveUploads, metrics.UploadSpeeds, metrics.EventsCounter, metrics.CurrentMessages)
	metrics.RegisterMetrics(m...)

	metrics.DefaultPoller.Start(ctx, pollInterval)
} // setupMetrics
