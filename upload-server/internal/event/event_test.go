package event

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
)

func TestBasicMemoryEvent(t *testing.T) {
	ctx := context.Background()
	c := make(chan *FileReady)

	pub := MemoryPublisher[*FileReady]{c}
	sub := MemorySubscriber[*FileReady]{c}

	var receivedEvent *FileReady
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		sub.Listen(ctx, func(ctx context.Context, fr *FileReady) error {
			receivedEvent = fr
			wg.Done()
			return nil
		})
	}()

	testEvent := FileReady{
		Event: Event{
			ID:   "1234",
			Type: FileReadyEventType,
		},
		UploadId: "test-upload-id",
	}
	pub.Publish(ctx, &testEvent)
	wg.Wait()

	if receivedEvent == nil {
		t.Fatalf("subscriber never received test event")
	}

	if receivedEvent.UploadId != testEvent.UploadId {
		t.Fatalf("unexpected event received; expected %+v; got %+v", testEvent, receivedEvent)
	}
}

func TestUploadIDLoggingProcessor(t *testing.T) {
	// init test logger
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logOutput, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	c := make(chan *FileReady)

	pub := MemoryPublisher[*FileReady]{c}
	sub := MemorySubscriber[*FileReady]{c}

	var receivedEvent *FileReady
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		sub.Listen(ctx, UploadIDLoggerProcessor(
			func(c context.Context, e *FileReady) error {
				ctx = c
				l := sloger.FromContext(c)
				l.Info("handling test event")
				receivedEvent = e
				wg.Done()
				return nil
			},
		))
	}()

	testEvent := FileReady{
		Event: Event{
			ID:   "1234",
			Type: FileReadyEventType,
		},
		UploadId: "test-upload-id",
	}
	pub.Publish(ctx, &testEvent)
	wg.Wait()

	if receivedEvent == nil {
		t.Fatalf("subscriber never received test event")
	}

	if receivedEvent.UploadId != testEvent.UploadId {
		t.Fatalf("unexpected event received; expected %+v; got %+v", testEvent, receivedEvent)
	}
	if ctx.Value(sloger.LoggerKey) == nil {
		t.Fatalf("expected logger to be set in context")
	}
	if !strings.Contains(logOutput.String(), testEvent.UploadId) {
		t.Fatalf("expected log output to contain %s, got %s", testEvent.UploadId, logOutput.String())
	}
}
