package serverdex

// Prometheus:
// - Gauges, can go up-down, mem usage, temp, measuring a numeric value that you want to expose
// - Counters, http requests
// - Summaries
// - Histograms

// Usage
//	collector := prometheuscollector.New()
//	prometheus.MustRegister(collector)

import (
	"sync/atomic"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
) // .import

var (
	uploadsCopiedTusToDex = prometheus.NewDesc(
		"dex_uploads_copied_tus_to_dex",
		"Number of copied uploads to DEX.",
		nil, nil)
	uploadsCopiedTusToRouter = prometheus.NewDesc(
		"dex_uploads_copied_tus_to_router",
		"Number of copied uploads to Router.",
		nil, nil)
	uploadsCopiedTusToEdav = prometheus.NewDesc(
		"dex_uploads_copied_tus_to_edav",
		"Number of copied uploads to Edav.",
		nil, nil)
) // .var

type Collector struct {
	metrics metrics.Metrics
} // .Collector

// New creates a new collector with the provided metrics struct
func NewCollector(metrics metrics.Metrics) Collector {
	return Collector{
		metrics: metrics,
	} // .return
} // .New

func (Collector) Describe(descs chan<- *prometheus.Desc) {

	descs <- uploadsCopiedTusToDex
	descs <- uploadsCopiedTusToRouter
	descs <- uploadsCopiedTusToEdav

} // .Describe

func (c Collector) Collect(metrics chan<- prometheus.Metric) {

	metrics <- prometheus.MustNewConstMetric(
		uploadsCopiedTusToDex,
		prometheus.CounterValue,
		float64(atomic.LoadUint64(c.metrics.CopiedUploadToDex)),
	) // .metrics

	metrics <- prometheus.MustNewConstMetric(
		uploadsCopiedTusToRouter,
		prometheus.CounterValue,
		float64(atomic.LoadUint64(c.metrics.CopiedUploadToRouter)),
	) // .metrics

	metrics <- prometheus.MustNewConstMetric(
		uploadsCopiedTusToEdav,
		prometheus.CounterValue,
		float64(atomic.LoadUint64(c.metrics.CopiedUploadToEdav)),
	) // .metrics

} // .Collect
