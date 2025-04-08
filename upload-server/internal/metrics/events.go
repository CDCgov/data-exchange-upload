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
	Length(ctx context.Context) (float64, error)
}

type QueuePoller struct {
	queueMap map[string]Countable
	t        *time.Ticker
}

var DefaultPoller = QueuePoller{
	queueMap: make(map[string]Countable),
}

func (qp *QueuePoller) Start(ctx context.Context, interval time.Duration) {
	if qp.t != nil {
		return
	}
	qp.t = time.NewTicker(interval)

	go func() {
		defer func() {
			qp.t.Stop()
			qp.t = nil
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-qp.t.C:
				for q, c := range qp.queueMap {
					l, err := c.Length(ctx)
					if err != nil {
						slog.Warn("failed to get queue length", "queue", q, "reason", err)
						continue
					}
					CurrentMessages.With(prometheus.Labels{"queue": q}).Set(l)
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
