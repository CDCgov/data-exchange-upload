package postprocessing

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventgrid"
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
	Client azeventgrid.Client
}

type EventProcessable interface {
	GetEventBatch(max int) []event.FileReadyEvent
	Process(ctx context.Context, event event.FileReadyEvent)
}

func (mw *MemoryEventListener) GetEventBatch(_ int) []event.FileReadyEvent {
	evt := <-mw.C
	return []event.FileReadyEvent{evt}
}

func (mw *MemoryEventListener) Process(ctx context.Context, e event.FileReadyEvent) {
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
