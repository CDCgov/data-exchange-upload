package metrics

import (
	"context"
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var Labels = struct {
	EventType string
	EventOp   string
	QueueURL  string
}{
	EventType: "type",
	EventOp:   "op",
	QueueURL:  "queue",
}

var EventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "dex_server_events_total",
	Help: "Number of events that have been created or handled by the server",
}, []string{Labels.EventType, Labels.EventOp})

var CurrentMessages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "dex_server_queue_messages",
	Help: "Current number of messages in an event queue",
}, []string{Labels.QueueURL})

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

func (qp *QueuePoller) Start(ctx context.Context, interval time.Duration) context.CancelFunc {
	c, cancel := context.WithCancel(ctx)

	if qp.t != nil {
		return cancel
	}
	qp.t = time.NewTicker(interval)
	go func(c context.Context) {
		defer func() {
			qp.t.Stop()
			qp.t = nil
		}()
		for {
			select {
			case <-c.Done():
				return
			case <-qp.t.C:
				for q, c := range qp.queueMap {
					l, err := c.Length(ctx)
					if err != nil {
						slog.Warn("failed to get queue length", "queue", q, "reason", err)
						continue
					}
					CurrentMessages.With(prometheus.Labels{Labels.QueueURL: q}).Set(l)
				}
			}
		}
	}(c)

	return cancel
}

func RegisterQueue(url string, q any) {
	if c, ok := q.(Countable); ok {
		DefaultPoller.queueMap[url] = c
		CurrentMessages.With(prometheus.Labels{Labels.QueueURL: url}).Set(0)
	} else {
		slog.Warn("metrics could not register queue", "queue", q)
	}
}
