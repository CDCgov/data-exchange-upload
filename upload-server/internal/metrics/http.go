package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var OpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "dex_server_connections_open",
	Help: "Current number of server open connections.",
}) // .metricsOpenConnections

var HttpReqs = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
	},
	[]string{"code", "method"},
)

type codedResponseWriter struct {
	http.ResponseWriter
	code string
}

func (c *codedResponseWriter) WriteHeader(statusCode int) {
	c.code = strconv.Itoa(statusCode)
	c.ResponseWriter.WriteHeader(statusCode)
}

func TrackHTTP(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		OpenConnections.Inc()
		ww := &codedResponseWriter{
			ResponseWriter: w,
		}
		handler.ServeHTTP(ww, r)
		HttpReqs.WithLabelValues(ww.code, r.Method).Inc()
		OpenConnections.Dec()
	})
}
