package event

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

var FileReadyChan chan *FileReady

func InitFileReadyChannel() {
	FileReadyChan = make(chan *FileReady, 100)
}

func CloseFileReadyChannel() {
	close(FileReadyChan)
}

func GetChannel[T Identifiable]() (chan T, error) {
	if r, ok := any(FileReadyChan).(chan T); ok {
		return r, nil
	}

	return nil, fmt.Errorf("channel not found")
}

type MemorySubscriber[T Identifiable] struct {
	Chan chan T
}

func (ms *MemorySubscriber[T]) Listen(ctx context.Context, process func(context.Context, T) error) error {
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

func (ms *MemorySubscriber[T]) Close() error {
	slog.Info("closing in-memory subscriber")
	return nil
}

func (ms *MemorySubscriber[T]) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Subscriber"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (ms *MemorySubscriber[T]) Length() int {
	return len(ms.Chan)
}

type MemoryPublisher[T Identifiable] struct {
	Chan chan T
}

func (mp *MemoryPublisher[T]) Publish(_ context.Context, event T) error {
	if mp.Chan != nil {
		go func() {
			mp.Chan <- event
		}()
		return nil
	}
	return errors.New("No event channel found")
}

func (mp *MemoryPublisher[T]) Close() error {
	return nil
}

func (mp *MemoryPublisher[T]) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Publisher"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}
