package event

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type Publishers[T Identifiable] []Publisher[T]

func (p Publishers[T]) Publish(ctx context.Context, e T) error {
	var errs error
	for _, publisher := range p {
		for range MaxRetries {
			if err := publisher.Publish(ctx, e); err != nil {
				slog.Error("Failed to publish", "event", e, "publisher", publisher, "err", err)
				errs = fmt.Errorf("Failed to publish event %s %w", e.Identifier(), err)
				continue
			}
			metrics.EventsCounter.With(prometheus.Labels{metrics.Labels.EventType: e.Type(), metrics.Labels.EventOp: "publish"}).Inc()
			slog.Info("published event", "event", e)
			break
		}
	}
	return errs
}

func (p Publishers[T]) Close() {
	for _, publisher := range p {
		c, ok := publisher.(io.Closer)
		if ok {
			c.Close()
		}
	}
}

type Publisher[T Identifiable] interface {
	Publish(ctx context.Context, event T) error
}
