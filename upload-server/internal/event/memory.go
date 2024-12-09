package event

import (
	"context"
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

type MemoryBus[T Identifiable] struct {
	Chan   chan T
	closed bool
}

func (ms *MemoryBus[T]) Listen(ctx context.Context, process func(context.Context, T) error) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case evt := <-ms.Chan:
			if err := process(ctx, evt); err != nil {
				slog.Error("failed to handle event", "event", evt, "error", err.Error())
				if evt.RetryCount() < MaxRetries {
					evt.IncrementRetryCount()
					// Retrying in a separate go routine so this doesn't block on channel write.
					go func() {
						ms.Chan <- evt
					}()
				}
			}
		}
	}
}

func (ms *MemoryBus[T]) Close() error {
	ms.closed = true
	if !ms.closed && ms.Chan != nil {
		close(ms.Chan)
	}
	return nil
}

func (ms *MemoryBus[T]) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Subscriber"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (mp *MemoryBus[T]) Publish(_ context.Context, event T) error {
	if mp.Chan != nil && !mp.closed {
		go func() {
			mp.Chan <- event
		}()
	}
	return nil
}
