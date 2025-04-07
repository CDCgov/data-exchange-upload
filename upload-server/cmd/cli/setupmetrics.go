package cli

// tusd metrics loading: https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
// note: tusd.Handler exposes metrics by cli flag and defaults true

import (
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tus/tusd/v2/pkg/hooks"
) // .import

// NOTE: pull out manifestMetrics and the dependency on appconfig goes away
func setupMetrics(m ...prometheus.Collector) {
	metrics.RegisterMetrics(hooks.MetricsHookErrorsTotal, hooks.MetricsHookInvocationsTotal, metrics.HttpReqs, metrics.OpenConnections, metrics.ActiveUploads, metrics.UploadSpeeds, metrics.QueuedDeliveries)
	metrics.RegisterMetrics(m...)
} // setupMetrics
