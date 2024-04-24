package cli

// tusd metrics loading: https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
// note: tusd.Handler exposes metrics by cli flag and defaults true

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tus/tusd/v2/pkg/hooks"
) // .import

func setupMetrics() {

	// ------------------------------------------------------------------
	// metrics as exposed by TUSD, refs:
	// https://github.com/tus/tusd/blob/main/cmd/tusd/cli/metrics.go
	// ------------------------------------------------------------------

	prometheus.MustRegister(hooks.MetricsHookErrorsTotal)
	prometheus.MustRegister(hooks.MetricsHookInvocationsTotal)

} // setupMetrics
