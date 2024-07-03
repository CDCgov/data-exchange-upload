package event

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/aznamespaces"
)

type MemoryEventSubscriber struct {
	C chan FileReadyEvent
}

type AzureEventSubscriber struct {
	Client aznamespaces.ReceiverClient
}

type Subscribable interface {
	GetBatch(ctx context.Context, max int) ([]FileReadyEvent, error)
	HandleSuccess(ctx context.Context, event FileReadyEvent) error
	HandleError(ctx context.Context, event FileReadyEvent, handlerError error)
}

func (mel *MemoryEventSubscriber) GetBatch(_ context.Context, _ int) ([]FileReadyEvent, error) {
	evt := <-mel.C
	return []FileReadyEvent{evt}, nil
}

func (mel *MemoryEventSubscriber) HandleSuccess(_ context.Context, e FileReadyEvent) error {
	logger.Info("successfully delivered file to target", "target", e.DeliverTarget)
	return nil
}

func (mel *MemoryEventSubscriber) HandleError(_ context.Context, e FileReadyEvent, err error) {
	logger.Error("failed to deliver file to target", "target", e.DeliverTarget, "error", err.Error())
}

func (ael *AzureEventSubscriber) GetBatch(ctx context.Context, max int) ([]FileReadyEvent, error) {
	resp, _ := ael.Client.ReceiveEvents(ctx, &aznamespaces.ReceiveEventsOptions{
		MaxEvents:   to.Ptr(int32(max)),
		MaxWaitTime: to.Ptr[int32](60),
	})

	var fileReadyEvents []FileReadyEvent
	for _, e := range resp.Details {
		logger.Info("received event", "event", e.Event.Data)

		var fre FileReadyEvent
		fre, err := NewFileReadyEventFromCloudEvent(e.Event, *e.BrokerProperties.LockToken)
		if err != nil {
			return nil, err
		}
		fileReadyEvents = append(fileReadyEvents, fre)
	}

	return fileReadyEvents, nil
}

func (ael *AzureEventSubscriber) HandleSuccess(ctx context.Context, e FileReadyEvent) error {
	_, err := ael.Client.AcknowledgeEvents(ctx, []string{e.Event.LockToken}, nil)
	if err != nil {
		logger.Error("failed to ack event", "error", err)
		return err
	}
	logger.Info("successfully handled event", "event", e)
	return nil
}

func (ael *AzureEventSubscriber) HandleError(ctx context.Context, e FileReadyEvent, handlerError error) {
	logger.Error("failed to handle event", "event", e, "error", handlerError.Error())
	resp, err := ael.Client.RejectEvents(ctx, []string{e.Event.LockToken}, nil)
	if err != nil {
		// TODO need to handle this better
		logger.Error("failed to reject events", "error", err.Error())
		for _, t := range resp.FailedLockTokens {
			logger.Error("failed to dead letter event with lock token", "token", t)
		}
	}

	for _, t := range resp.SucceededLockTokens {
		logger.Info("successfully dead lettered event with lock token", "token", t)
	}
}
