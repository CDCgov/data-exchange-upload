package event

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/messaging"
)

type Event struct {
	ID        string
	Type      string
	LockToken string
}

type FileReadyEvent struct {
	Event
	Manifest      map[string]string
	DeliverTarget string
}

func NewFileReadyEvent(id string, manifest map[string]string, target string) FileReadyEvent {
	return FileReadyEvent{
		Event: Event{
			ID:   id,
			Type: FileReadyEventType,
		},
		Manifest:      manifest,
		DeliverTarget: target,
	}
}

func NewFileReadyEventFromCloudEvent(event messaging.CloudEvent, lockToken string) (FileReadyEvent, error) {
	var fre FileReadyEvent
	err := json.Unmarshal(event.Data.([]byte), &fre)
	if err != nil {
		return fre, err
	}

	fre.LockToken = lockToken

	return fre, nil
}
