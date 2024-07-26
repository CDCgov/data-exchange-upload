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
	UploadId          string `json:"upload_id"`
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

func NewFileReadyEvent(uploadId string, metadata map[string]string, target string) FileReady {
	return FileReady{
		Event: Event{
			Type: FileReadyEventType,
		},
		UploadId:          uploadId,
		Metadata:          metadata,
		DestinationTarget: target,
	}
}

func NewFileReadyEventFromServiceBusMessage(m azservicebus.ReceivedMessage) (FileReady, error) {
	var fre FileReady
	err := json.Unmarshal(m.Body, &fre)
	if err != nil {
		return fre, err
	}

	fre.ID = m.MessageID
	fre.OriginalMessage = m

	return fre, nil
}
