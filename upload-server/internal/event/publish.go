package event

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/messaging"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/aznamespaces"
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

type Event struct {
	ID string
}

type FileReadyEvent struct {
	Event
	Manifest      map[string]string
	DeliverTarget string
}

type Publisher interface {
	Publish(ctx context.Context, event FileReadyEvent) error
}

type MemoryPublisher struct {
	FileReadyChannel chan FileReadyEvent
}

type AzurePublisher struct {
	Client *aznamespaces.SenderClient
}

func (mp *MemoryPublisher) Publish(_ context.Context, event FileReadyEvent) error {
	mp.FileReadyChannel <- event
	return nil
}

// TODO batch events
func (ap *AzurePublisher) Publish(ctx context.Context, event FileReadyEvent) error {
	logger.Info("publishing file ready event")
	evt, err := messaging.NewCloudEvent("upload", "fileReady", event, nil)
	if err != nil {
		return err
	}

	_, err = ap.Client.SendEvent(ctx, &evt, nil)

	return err
}
