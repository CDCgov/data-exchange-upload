package metrics

import (
	"errors"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

// todo this could also be a vec per datastream
var ActiveUploads = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "dex_server_active_uploads",
	Help: "Current number of active uploads",
}) // .metricsOpenConnections

var UploadDurationSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name: "dex_server_upload_duration_seconds",
	Help: "File upload duration distribution in seconds",
	// TODO parameterize this with a helper function; can eventually be driven by config.
	Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 180, 240, 300, 450, 600},
})

var DefaultMetrics = []prometheus.Collector{
	ActiveUploads,
	UploadDurationSeconds,
}

func ActiveUploadIncHook(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	ActiveUploads.Inc()
	return resp, nil
}
func ActiveUploadDecHook(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	ActiveUploads.Dec()
	return resp, nil
}

func UploadDurationObserveHook(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	tuid := event.Upload.ID
	if resp.ChangeFileInfo.ID != "" {
		tuid = resp.ChangeFileInfo.ID
	}
	if tuid == "" {
		return resp, errors.New("no Upload ID defined")
	}

	manifest := event.Upload.MetaData
	start, ok := manifest["dex_ingest_datetime"]
	if !ok {
		slog.Warn("unable to observe upload duration; no start time found in manifest", "uploadId", tuid)
		return resp, nil
	}

	startTime, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		slog.Warn("unable to observe upload duration; unable to parse timestamp", "timestamp", start, "uploadId", tuid)
		return resp, nil
	}

	duration := time.Since(startTime).Seconds()
	slog.Info("observed upload duration", "duration", duration)
	UploadDurationSeconds.Observe(duration)

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
				val = ""
			}
		}
		vals = append(vals, val)
	}
	mm.Counter.WithLabelValues(vals...).Inc()
	return resp, nil
}
