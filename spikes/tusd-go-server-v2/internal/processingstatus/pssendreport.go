package processingstatus

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/models"
)

func (pss PsSender) SendReport(report models.PsReport) error { // TODO: probably not return if an error

	// TODO: what happens when we can't send to processing status and error?
	// should it even be returned? probably not so this can be called on a goroutine

	sender, err := pss.ServiceBusClient.NewSender(pss.ServiceBusQueue, nil)
	if err != nil {
		return err
	} // .if
	defer sender.Close(context.TODO())

	jsonData, err := json.Marshal(report)
	if err != nil {
		return err
	} // .if

	sbMessage := &azservicebus.Message{

		Body: []byte(jsonData),
	} // .sbMessage
	err = sender.SendMessage(context.TODO(), sbMessage, nil)
	if err != nil {
		return err
	} // .if

	return nil // all good no errors
} // .Send
