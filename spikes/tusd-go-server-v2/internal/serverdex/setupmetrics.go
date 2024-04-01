package serverdex

// tusd metrics loading: https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
// note: tusd.Handler exposes metrics by cli flag and defaults true

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tus/tusd/v2/pkg/hooks"
) // .import

var metricsOpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "dex_server_connections_open",
	Help: "Current number of server open connections.",
}) // .metricsOpenConnections

func (sd ServerDex) setupMetrics() {

	// ------------------------------------------------------------------
	// metrics as exposed by TUSD, refs:
	// https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
	// ------------------------------------------------------------------

	prometheus.MustRegister(metricsOpenConnections)
	prometheus.MustRegister(hooks.MetricsHookErrorsTotal)
	prometheus.MustRegister(hooks.MetricsHookInvocationsTotal)

	// ------------------------------------------------------------------
	// DEX server metrics
	// ------------------------------------------------------------------

	prometheus.MustRegister(NewCollector(*metrics.DefaultMetrics))
} // setupMetrics
