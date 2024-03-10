package serverdex

// tusd metrics loading: https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
// note: tusd.Handler exposes metrics by cli flag and defaults true

import (
	"github.com/prometheus/client_golang/prometheus"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	"github.com/tus/tusd/v2/pkg/prometheuscollector"
) // .import

var MetricsOpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "tusd_connections_open",
	Help: "Current number of open connections.",
})

func (sd ServerDex) setupMetrics(handlerTusd *tusd.Handler) {

	prometheus.MustRegister(MetricsOpenConnections)
	prometheus.MustRegister(hooks.MetricsHookErrorsTotal)
	prometheus.MustRegister(hooks.MetricsHookInvocationsTotal)
	prometheus.MustRegister(prometheuscollector.New(handlerTusd.Metrics))
} // setupMetrics
