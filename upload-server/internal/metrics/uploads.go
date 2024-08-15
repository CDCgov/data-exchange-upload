package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

// todo this could also be a vec per datastream
var activeUploads = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "dex_server_active_uploads",
	Help: "Current number of active uploads",
}) // .metricsOpenConnections

var DefaultMetrics = []prometheus.Collector{
	activeUploads,
}

func ActiveUploadIncHook(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	activeUploads.Inc()
	return resp, nil
}
func ActiveUploadDecHook(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	activeUploads.Dec()
	return resp, nil
}

func RegisterMetrics(metrics ...prometheus.Collector) error {
	if metrics == nil {
		metrics = DefaultMetrics
	}
	for _, m := range metrics {
		if err := prometheus.Register(m); err != nil {
			return err
		}
	}
	return nil
}

func NewManifestMetrics(name string, help string, keys ...string) *ManifestMetrics {
	// todo: structure this to make it work with a config for manifest fields
	c := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
			Help: help,
		},
		keys,
	)
	return &ManifestMetrics{
		Keys:    keys,
		Counter: c,
	}
}

type ManifestMetrics struct {
	Keys    []string
	Counter *prometheus.CounterVec
}

func (mm *ManifestMetrics) Hook(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	vals := []string{}
	for _, key := range mm.Keys {
		val, ok := event.Upload.MetaData[key]
		if !ok {
			val, ok = resp.ChangeFileInfo.MetaData[key]
			if !ok {
				// no op.. should be something else
				continue
			}
		}
		vals = append(vals, val)
	}
	mm.Counter.WithLabelValues(vals...).Inc()
	return resp, nil
}
