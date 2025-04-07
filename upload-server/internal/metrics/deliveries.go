package metrics

import "github.com/prometheus/client_golang/prometheus"

var QueuedDeliveries = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "dex_server_delivery_queue_size",
	Help: "Current number of uploads that are queued for deliveries",
})
