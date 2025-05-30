package metrics

import (
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

const MB = 1024 * 1024

// todo this could also be a vec per datastream
var ActiveUploads = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "dex_server_active_uploads",
	Help: "Current number of active uploads",
}) // .metricsOpenConnections

var UploadSpeeds = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "dex_server_upload_speed_bytes_per_second",
	Help:    "File upload speed distribution",
	Buckets: prometheus.ExponentialBuckets(10, 2.5, 20),
})

var DefaultMetrics = []prometheus.Collector{
	ActiveUploads,
	UploadSpeeds,
}

func ActiveUploadIncHook(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	ActiveUploads.Inc()
	return resp, nil
}
func ActiveUploadDecHook(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	ActiveUploads.Dec()
	return resp, nil
}

func UploadSpeedsHook(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	logger := sloger.FromContext(event.Context)

	size := event.Upload.Size

	manifest := event.Upload.MetaData
	start, ok := manifest["dex_ingest_datetime"]
	if !ok {
		logger.Warn("unable to observe upload duration; no start time found in manifest")
		return resp, nil
	}

	startTime, err := time.Parse(time.RFC3339Nano, start)
	if err != nil {
		logger.Warn("unable to observe upload duration; unable to parse timestamp", "timestamp", start)
		return resp, nil
	}

	duration := time.Since(startTime).Seconds()
	if duration > 0 {
		speed := float64(size) / duration
		UploadSpeeds.Observe(speed)
	}

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

func (mm *ManifestMetrics) Hook(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
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
