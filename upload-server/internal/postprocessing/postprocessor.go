package postprocessing

import (
	"context"
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/aznamespaces"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"log/slog"
	"reflect"
	"strings"
)

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

type PostProcessor struct {
	UploadBaseDir string
	UploadDir     string
}

type MemoryEventListener struct {
	C chan event.FileReadyEvent
}

type AzureEventListener struct {
	Client aznamespaces.ReceiverClient
}

func ProcessFileReadyEvent(ctx context.Context, e event.FileReadyEvent) error {
	return Deliver(ctx, e.ID, e.Manifest, e.DeliverTarget)
}

type EventProcessable interface {
	GetEventBatch(ctx context.Context, max int) ([]event.FileReadyEvent, error)
	HandleSuccess(ctx context.Context, event event.FileReadyEvent) error
	HandleError(ctx context.Context, event event.FileReadyEvent, handlerError error)
}

func (mel *MemoryEventListener) GetEventBatch(_ context.Context, _ int) []event.FileReadyEvent {
	evt := <-mel.C
	return []event.FileReadyEvent{evt}
}

func (mel *MemoryEventListener) HandleSuccess(_ context.Context, e event.FileReadyEvent) {
	logger.Info("successfully delivered file to target", "target", e.DeliverTarget)
}

func (mel *MemoryEventListener) HandleError(_ context.Context, e event.FileReadyEvent, err error) {
	logger.Error("failed to deliver file to target", "target", e.DeliverTarget, "error", err.Error())
}

func (ael *AzureEventListener) GetEventBatch(ctx context.Context, max int) ([]event.FileReadyEvent, error) {
	resp, _ := ael.Client.ReceiveEvents(ctx, &aznamespaces.ReceiveEventsOptions{
		MaxEvents:   to.Ptr(int32(max)),
		MaxWaitTime: to.Ptr[int32](60),
	})

	var fileReadyEvents []event.FileReadyEvent
	for _, e := range resp.Details {
		logger.Info("received event", "event", e.Event.Data)

		var fre event.FileReadyEvent
		err := json.Unmarshal(e.Event.Data.([]byte), &fre)
		if err != nil {
			return nil, err
		}
		fileReadyEvents = append(fileReadyEvents, fre)
	}

	return fileReadyEvents, nil
}

func (ael *AzureEventListener) HandleSuccess(ctx context.Context, e event.FileReadyEvent) error {
	_, err := ael.Client.AcknowledgeEvents(ctx, []string{e.Event.LockToken}, nil)
	if err != nil {
		logger.Error("failed to ack event", "error", err)
		return err
	}
	return nil
}

func (ael *AzureEventListener) HandleError(ctx context.Context, e event.FileReadyEvent, handlerError error) {
	logger.Error("failed to handle event", "event", e, "error", handlerError.Error())
	_, err := ael.Client.RejectEvents(ctx, []string{e.Event.LockToken}, nil)
	if err != nil {
		logger.Error("failed to reject events", "error", err.Error())
	}
}

func (ael *AzureEventListener) Process(ctx context.Context, e event.FileReadyEvent) {
	if err := Deliver(ctx, e.ID, e.Manifest, e.DeliverTarget); err != nil {
		// TODO Retry
		logger.Error("error delivering file to target", "event", e, "error", err.Error())
	}
}

type Event struct {
	ID       string
	Manifest map[string]string
	Target   string
}

func Worker(ctx context.Context, c chan event.FileReadyEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-c:
			if err := Deliver(ctx, e.ID, e.Manifest, e.DeliverTarget); err != nil {
				// TODO Retry
				logger.Error("error delivering file to target", "event", e, "error", err.Error())
			}
		}
	}
}
