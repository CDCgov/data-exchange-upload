package metrics

import "github.com/prometheus/client_golang/prometheus"

var EventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dex_server_events_total",
	Help: "Number of file ready events that have been enqueued for files ready for delivery",
}, []string{"type", "op"})
