package event

import (
	"encoding/json"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

const FileReadyEventType = "FileReady"

var FileReadyChan chan *FileReady

// TODO better name for this interface would be Subscribable or Queueable or similar
type Identifiable interface {
	Identifier() string
	Type() string
	OrigMessage() *azservicebus.ReceivedMessage
	SetIdentifier(id string)
	SetType(t string)
	SetOrigMessage(m *azservicebus.ReceivedMessage)
}

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
	OriginalMessage   *azservicebus.ReceivedMessage `json:"-"`
}

func (fr *FileReady) Type() string {
	return fr.Event.Type
}

func (fr *FileReady) OrigMessage() *azservicebus.ReceivedMessage {
	return fr.OriginalMessage
}

func (fr *FileReady) SetIdentifier(id string) {
	fr.ID = id
}

func (fr *FileReady) SetType(t string) {
	fr.Event.Type = t
}

func (fr *FileReady) SetOrigMessage(m *azservicebus.ReceivedMessage) {
	fr.OriginalMessage = m
}

func (fr *FileReady) Identifier() string {
	return fr.UploadId
}

func InitFileReadyChannel() {
	FileReadyChan = make(chan *FileReady)
}

func CloseFileReadyChannel() {
	close(FileReadyChan)
}

func NewFileReadyEvent(uploadId string, metadata map[string]string, target string) *FileReady {
	return &FileReady{
		Event: Event{
			Type: FileReadyEventType,
		},
		UploadId:          uploadId,
		Metadata:          metadata,
		DestinationTarget: target,
	}
}

func NewEventFromServiceBusMessage[T Identifiable](m *azservicebus.ReceivedMessage) (T, error) {
	logger.Info("casting event ***", "event", m)
	var e T
	err := json.Unmarshal(m.Body, &e)
	if err != nil {
		return e, err
	}

	e.SetIdentifier(m.MessageID)
	e.SetOrigMessage(m)

	return e, nil
}
