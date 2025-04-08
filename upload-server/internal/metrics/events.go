package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var EventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dex_server_events_total",
	Help: "Number of file ready events that have been enqueued for files ready for delivery",
}, []string{"queue", "op"})

var CurrentMessages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "dex_server_queue_messages",
	Help: "Current number of messages in an event queue",
}, []string{"queue"})

type Countable interface {
	Length() int
}

type QueuePoller struct {
	queueMap map[string]Countable
	t        *time.Ticker
}

var DefaultPoller = QueuePoller{
	queueMap: make(map[string]Countable),
}

func (qp *QueuePoller) Start(ctx context.Context) {
	qp.t = time.NewTicker(500 * time.Millisecond)

	go func() {
		defer qp.t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-qp.t.C:
				for q, c := range qp.queueMap {
					CurrentMessages.With(prometheus.Labels{"queue": q}).Set(float64(c.Length()))
				}
			}
		}
	}()
}

func RegisterQueue(name string, q any) {
	if c, ok := q.(Countable); ok {
		DefaultPoller.queueMap[name] = c
	} else {
		slog.Warn("metrics could not register queue", "queue", q)
	}
}
