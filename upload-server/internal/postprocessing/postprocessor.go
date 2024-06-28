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

type EventProcessable interface {
	GetEventBatch(ctx context.Context, max int) []event.FileReadyEvent
	Process(ctx context.Context, event event.FileReadyEvent) // TODO separate out.  Probs can be top level func
}

func (mel *MemoryEventListener) GetEventBatch(_ context.Context, _ int) []event.FileReadyEvent {
	evt := <-mel.C
	return []event.FileReadyEvent{evt}
}

func (mel *MemoryEventListener) Process(ctx context.Context, e event.FileReadyEvent) {
	if err := Deliver(ctx, e.ID, e.Manifest, e.DeliverTarget); err != nil {
		// TODO Retry
		logger.Error("error delivering file to target", "event", e, "error", err.Error())
	}
	//for {
	//	select {
	//	case <-ctx.Done():
	//		return
	//	case e := <-mw.C:
	//		if err := Deliver(ctx, e.ID, e.Manifest, e.DeliverTarget); err != nil {
	//			// TODO Retry
	//			logger.Error("error delivering file to target", "event", e, "error", err.Error())
	//		}
	//	}
	//}
}

func (ael *AzureEventListener) GetEventBatch(ctx context.Context, max int) []event.FileReadyEvent {
	resp, _ := ael.Client.ReceiveEvents(ctx, &aznamespaces.ReceiveEventsOptions{
		MaxEvents:   to.Ptr(int32(max)),
		MaxWaitTime: to.Ptr[int32](60),
	})
	// TODO error handling
	// TODO Covert cloud event to file ready event
	var fileReadyEvents []event.FileReadyEvent
	for _, e := range resp.Details {
		logger.Info("received event", "event", e.Event.Data)

		// TODO type check
		var fre event.FileReadyEvent
		json.Unmarshal(e.Event.Data.([]byte), &fre)
		fileReadyEvents = append(fileReadyEvents, fre)
		// TODO move to process
		_, err := ael.Client.AcknowledgeEvents(ctx, []string{*e.BrokerProperties.LockToken}, nil)
		if err != nil {
			logger.Error("failed to ack event", "error", err)
		}
	}

	return fileReadyEvents
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
