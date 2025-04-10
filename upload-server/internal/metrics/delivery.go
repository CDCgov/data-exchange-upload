package metrics

import "github.com/prometheus/client_golang/prometheus"

var ActiveDeliveries = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "dex_server_active_deliveries",
	Help: "Gauge showing number of deliveries in progress",
}, []string{"target"})

var DeliveryTotals = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dex_server_deliveries_total",
	Help: "Number of deliveries that have been handled by the server",
}, []string{"target", "result"})
