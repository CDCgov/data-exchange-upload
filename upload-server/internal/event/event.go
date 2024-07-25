package event

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

const FileReadyEventType = "FileReady"

var fileReadyChan chan FileReady

type Event struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type FileReady struct {
	Event
	SrcUrl            string `json:"src_url"`
	DestinationTarget string `json:"deliver_target"`
	Metadata          map[string]string
	OriginalMessage   azservicebus.ReceivedMessage
}

func InitFileReadyChannel() {
	fileReadyChan = make(chan FileReady)
}

func CloseFileReadyChannel() {
	close(fileReadyChan)
}

func NewFileReadyEvent(id string, metadata map[string]string, target string) FileReady {
	return FileReady{
		Event: Event{
			ID:   id,
			Type: FileReadyEventType,
		},
		Metadata:          metadata,
		DestinationTarget: target,
	}
}

//func NewFileReadyEventFromCloudEvent(event messaging.CloudEvent, lockToken string) (FileReady, error) {
//	var fre FileReady
//	err := json.Unmarshal(event.Data.([]byte), &fre)
//	if err != nil {
//		return fre, err
//	}
//
//	fre.LockToken = lockToken
//
//	return fre, nil
//}

func NewFileReadyEventFromServiceBusMessage(m azservicebus.ReceivedMessage) (FileReady, error) {
	var fre FileReady
	err := json.Unmarshal(m.Body, &fre)
	if err != nil {
		return fre, err
	}

	fre.ID = m.MessageID
	fre.Type = FileReadyEventType
	fre.OriginalMessage = m

	return fre, nil
}
