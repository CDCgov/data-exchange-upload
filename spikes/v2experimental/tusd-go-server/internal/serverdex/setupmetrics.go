package serverdex

// tusd metrics loading: https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
// note: tusd.Handler exposes metrics by cli flag and defaults true

import (
	"github.com/prometheus/client_golang/prometheus"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/prometheuscollector"
) // .import

var metricsOpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "tusd_connections_open",
	Help: "Current number of open connections.",
})

func (sd ServerDex) setupMetrics(handlerTusd *tusd.Handler) {

	// ------------------------------------------------------------------
	// metrics as exposed by TUSD, refs:
	// https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
	// ------------------------------------------------------------------

	prometheus.MustRegister(metricsOpenConnections)
	prometheus.MustRegister(hooks.MetricsHookErrorsTotal)
	prometheus.MustRegister(hooks.MetricsHookInvocationsTotal)
	prometheus.MustRegister(prometheuscollector.New(handlerTusd.Metrics))

	// ------------------------------------------------------------------
	// DEX server metrics
	// ------------------------------------------------------------------

	prometheus.MustRegister(NewCollector(sd.Metrics))
} // setupMetrics
