package processingstatus

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

func (pss PsSender) SendReport(message string) error { // TODO: probably not return if an error

	// TODO: what happens when we can't send to processing status and error?
	// should it even be returned? probably not so this can be called on a goroutine

	sender, err := pss.ServiceBusClient.NewSender(pss.ServiceBusQueue, nil)
	if err != nil {
		return err
	} // .if
	defer sender.Close(context.TODO())

	sbMessage := &azservicebus.Message{

		Body: []byte(message), // TODO change to .json report

	} // .sbMessage
	err = sender.SendMessage(context.TODO(), sbMessage, nil)
	if err != nil {
		return err
	} // .if

	return nil // all good no errors
} // .Send
