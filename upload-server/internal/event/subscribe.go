package event

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

type MemorySubscriber struct{}

type AzureSubscriber struct {
	EventType   string
	Receiver    azservicebus.Receiver
	Config      appconfig.AzureQueueConfig
	AdminClient admin.Client
}

type Subscribable interface {
	health.Checkable
	GetBatch(ctx context.Context, max int) ([]FileReady, error)
	HandleSuccess(ctx context.Context, event FileReady) error
	HandleError(ctx context.Context, event FileReady, handlerError error)
}

func (ms *MemorySubscriber) GetBatch(ctx context.Context, _ int) ([]FileReady, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case evt := <-fileReadyChan:
		return []FileReady{evt}, nil
	}
}

func (ms *MemorySubscriber) HandleSuccess(_ context.Context, e FileReady) error {
	logger.Info("successfully delivered file to target", "target", e.DestinationTarget)
	return nil
}

func (ms *MemorySubscriber) HandleError(_ context.Context, e FileReady, err error) {
	logger.Error("failed to deliver file to target", "target", e.DestinationTarget, "error", err.Error())
}

func (ms *MemorySubscriber) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Subscriber"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (as *AzureSubscriber) GetBatch(ctx context.Context, max int) ([]FileReady, error) {
	msgs, err := as.Receiver.ReceiveMessages(ctx, max, nil)
	if err != nil {
		return nil, err
	}

	//resp, _ := as.Client.ReceiveEvents(ctx, &aznamespaces.ReceiveEventsOptions{
	//	MaxEvents:   to.Ptr(int32(max)),
	//	MaxWaitTime: to.Ptr[int32](60),
	//})

	var fileReadyEvents []FileReady
	for _, m := range msgs {
		logger.Info("received event", "event", m.Body)

		var fre FileReady
		fre, err := NewFileReadyEventFromServiceBusMessage(*m)
		if err != nil {
			return nil, err
		}
		fileReadyEvents = append(fileReadyEvents, fre)
	}

	return fileReadyEvents, nil
}

func (as *AzureSubscriber) HandleSuccess(ctx context.Context, e FileReady) error {
	err := as.Receiver.CompleteMessage(ctx, &e.OriginalMessage, nil)
	//_, err := as.Client.AcknowledgeEvents(ctx, []string{e.Event.LockToken}, nil)
	if err != nil {
		logger.Error("failed to ack event", "error", err)
		return err
	}
	logger.Info("successfully handled event", "event", e)
	return nil
}

func (as *AzureSubscriber) HandleError(_ context.Context, e FileReady, handlerError error) {
	logger.Error("failed to handle event", "event", e, "error", handlerError.Error())
	// TODO dead letter message
	//resp, err := as.Client.RejectEvents(ctx, []string{e.Event.LockToken}, nil)
	//if err != nil {
	//	// TODO need to handle this better
	//	logger.Error("failed to reject events", "error", err.Error())
	//	for _, t := range resp.FailedLockTokens {
	//		logger.Error("failed to dead letter event with lock token", "token", t)
	//	}
	//}
	//
	//for _, t := range resp.SucceededLockTokens {
	//	logger.Info("successfully dead lettered event with lock token", "token", t)
	//}
}

func (as *AzureSubscriber) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
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

	//if as.Client == nil {
	//	rsp.Status = models.STATUS_DOWN
	//	rsp.HealthIssue = "Azure event subscriber not configured"
	//	return rsp
	//}
	//
	//// Check via management API
	//cred, err := azidentity.NewDefaultAzureCredential(nil)
	//if err != nil {
	//	rsp.Status = models.STATUS_DOWN
	//	rsp.HealthIssue = "Failed to authenticate to Azure"
	//}
	//_, err = armeventgrid.NewClientFactory(as.Config.Subscription, cred, nil)
	//if err != nil {
	//	rsp.Status = models.STATUS_DOWN
	//	rsp.HealthIssue = fmt.Sprintf("Failed to connect to namespace %s", as.Config.Endpoint)
	//}

	return rsp
}
