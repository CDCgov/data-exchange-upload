package metrics

import "github.com/prometheus/client_golang/prometheus"

var ActiveDeliveries = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "dex_server_active_deliveries",
	Help: "Gauge showing number of deliveries in progress",
}, []string{"target"})
