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

	"github.com/prometheus/client_golang/prometheus"
) // .import

var (
	uploadsCopiedAToB = prometheus.NewDesc(
		"dex_uploads_copied_a_to_b",
		"Number of copied uploads A to B.",
		nil, nil)
	uploadsCopiedBToC = prometheus.NewDesc(
		"dex_uploads_copied_b_to_c",
		"Number of copied uploads A to B.",
		nil, nil)
) // .var

type Collector struct {
	metrics Metrics
} // .Collector

// New creates a new collector with the provided metrics struct
func NewCollector(metrics Metrics) Collector {
	return Collector{
		metrics: metrics,
	} // .return
} // .New

func (Collector) Describe(descs chan<- *prometheus.Desc) {

	descs <- uploadsCopiedAToB
	descs <- uploadsCopiedBToC

} // .Describe

func (c Collector) Collect(metrics chan<- prometheus.Metric) {

	metrics <- prometheus.MustNewConstMetric(
		uploadsCopiedAToB,
		prometheus.CounterValue,
		float64(atomic.LoadUint64(c.metrics.CopiedAToB)),
	) // .metrics

	metrics <- prometheus.MustNewConstMetric(
		uploadsCopiedBToC,
		prometheus.CounterValue,
		float64(atomic.LoadUint64(c.metrics.CopiedBToC)),
	) // .metrics

} // .Collect
