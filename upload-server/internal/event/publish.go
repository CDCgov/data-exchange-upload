package event

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/messaging"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventgrid"
)

type Event struct {
	ID string
}

type FileReadyEvent struct {
	Event
	Manifest      map[string]string
	DeliverTarget string
}

type Publisher interface {
	Publish(ctx context.Context, event FileReadyEvent)
}

type MemoryPublisher struct {
	FileReadyChannel chan FileReadyEvent
}

type AzurePublisher struct {
	Client azeventgrid.Client
}

func (mp *MemoryPublisher) Publish(_ context.Context, event FileReadyEvent) {
	mp.FileReadyChannel <- event
}

// TODO batch events
func (ap *AzurePublisher) Publish(ctx context.Context, event FileReadyEvent) error {
	var eventsToSend []messaging.CloudEvent

	evt, err := messaging.NewCloudEvent("Upload API", "File Ready", event, nil)
	if err != nil {
		return err
	}

	eventsToSend = append(eventsToSend, evt)

	// TODO env var for topic name
	_, err = ap.Client.PublishCloudEvents(ctx, "upload-file-ready", eventsToSend, nil)

	return err
}
