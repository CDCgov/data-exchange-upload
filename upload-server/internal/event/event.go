package event

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/messaging"
)

type Event struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	LockToken string `json:"lock_token"`
}

type FileReady struct {
	Event
	Manifest      map[string]string `json:"manifest"`
	DeliverTarget string            `json:"deliver_target"`
}

func NewFileReadyEvent(id string, manifest map[string]string, target string) FileReady {
	return FileReady{
		Event: Event{
			ID:   id,
			Type: FileReadyEventType,
		},
		Manifest:      manifest,
		DeliverTarget: target,
	}
}

func NewFileReadyEventFromCloudEvent(event messaging.CloudEvent, lockToken string) (FileReady, error) {
	var fre FileReady
	err := json.Unmarshal(event.Data.([]byte), &fre)
	if err != nil {
		return fre, err
	}

	fre.LockToken = lockToken

	return fre, nil
}
