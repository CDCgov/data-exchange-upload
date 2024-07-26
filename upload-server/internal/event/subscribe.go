package event

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"io"
)

type MemorySubscriber[T Identifiable] struct {
	Chan chan T
}

type AzureSubscriber[T Identifiable] struct {
	Context     context.Context
	EventType   string
	Receiver    *azservicebus.Receiver
	Config      appconfig.AzureQueueConfig
	AdminClient *admin.Client
}

type Subscribable[T Identifiable] interface {
	health.Checkable
	io.Closer
	GetBatch(ctx context.Context, max int) ([]T, error)
	HandleSuccess(ctx context.Context, event T) error
	HandleError(ctx context.Context, event T, handlerError error)
}

func (ms *MemorySubscriber[T]) GetBatch(ctx context.Context, _ int) ([]T, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case evt := <-ms.Chan:
		return []T{evt}, nil
	}
}

func (ms *MemorySubscriber[T]) HandleSuccess(_ context.Context, e T) error {
	logger.Info("successfully handled event", "event", e)
	return nil
}

func (ms *MemorySubscriber[T]) HandleError(_ context.Context, e T, err error) {
	logger.Error("failed to handle event", "event", e, "error", err.Error())
}

func (ms *MemorySubscriber[T]) Close() error {
	logger.Info("closing in-memory subscriber")
	return nil
}

func (ms *MemorySubscriber[T]) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Subscriber"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (as *AzureSubscriber[T]) GetBatch(ctx context.Context, max int) ([]T, error) {
	msgs, err := as.Receiver.ReceiveMessages(ctx, max, nil)
	if err != nil {
		return nil, err
	}

	var batch []T
	for _, m := range msgs {
		logger.Info("received event", "event", m.Body)

		var e T
		e, err := NewEventFromServiceBusMessage[T](*m)
		if err != nil {
			return nil, err
		}
		batch = append(batch, e)
	}

	return batch, nil
}

func (as *AzureSubscriber[T]) HandleSuccess(ctx context.Context, e T) error {
	err := as.Receiver.CompleteMessage(ctx, e.OrigMessage(), nil)
	if err != nil {
		logger.Error("failed to ack event", "error", err)
		return err
	}
	logger.Info("successfully handled event", "event ID", e.Identifier(), "event type", e.Type())
	return nil
}

func (as *AzureSubscriber[T]) HandleError(_ context.Context, e T, handlerError error) {
	logger.Error("failed to handle event", "event ID", e.Identifier(), "event type", e.Type(), "error", handlerError.Error())
	// TODO dead letter message
}

func (as *AzureSubscriber[T]) Close() error {
	return as.Receiver.Close(as.Context)
}

func (as *AzureSubscriber[T]) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = fmt.Sprintf("%s Event Subscriber", as.EventType)
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE

	subResp, err := as.AdminClient.GetSubscription(ctx, as.Config.Topic, as.Config.Subscription, nil)
	if err != nil {
		return rsp.BuildErrorResponse(err)
	}

	if *subResp.Status != admin.EntityStatusActive {
		return rsp.BuildErrorResponse(fmt.Errorf("service bus subscription %s status: %s", as.Config.Subscription, *subResp.Status))
	}

	return rsp
}
