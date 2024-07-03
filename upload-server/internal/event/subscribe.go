package event

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/eventgrid/aznamespaces"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid/v2"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

type MemorySubscriber struct{}

type AzureSubscriber struct {
	Client *aznamespaces.ReceiverClient
	Config appconfig.AzureQueueConfig
}

type Subscribable interface {
	health.Checkable
	GetBatch(ctx context.Context, max int) ([]FileReady, error)
	HandleSuccess(ctx context.Context, event FileReady) error
	HandleError(ctx context.Context, event FileReady, handlerError error)
}

func (ms *MemorySubscriber) GetBatch(_ context.Context, _ int) ([]FileReady, error) {
	evt := <-fileReadyChan
	return []FileReady{evt}, nil
}

func (ms *MemorySubscriber) HandleSuccess(_ context.Context, e FileReady) error {
	logger.Info("successfully delivered file to target", "target", e.DeliverTarget)
	return nil
}

func (ms *MemorySubscriber) HandleError(_ context.Context, e FileReady, err error) {
	logger.Error("failed to deliver file to target", "target", e.DeliverTarget, "error", err.Error())
}

func (ms *MemorySubscriber) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Subscriber"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (as *AzureSubscriber) GetBatch(ctx context.Context, max int) ([]FileReady, error) {
	resp, _ := as.Client.ReceiveEvents(ctx, &aznamespaces.ReceiveEventsOptions{
		MaxEvents:   to.Ptr(int32(max)),
		MaxWaitTime: to.Ptr[int32](60),
	})

	var fileReadyEvents []FileReady
	for _, e := range resp.Details {
		logger.Info("received event", "event", e.Event.Data)

		var fre FileReady
		fre, err := NewFileReadyEventFromCloudEvent(e.Event, *e.BrokerProperties.LockToken)
		if err != nil {
			return nil, err
		}
		fileReadyEvents = append(fileReadyEvents, fre)
	}

	return fileReadyEvents, nil
}

func (as *AzureSubscriber) HandleSuccess(ctx context.Context, e FileReady) error {
	_, err := as.Client.AcknowledgeEvents(ctx, []string{e.Event.LockToken}, nil)
	if err != nil {
		logger.Error("failed to ack event", "error", err)
		return err
	}
	logger.Info("successfully handled event", "event", e)
	return nil
}

func (as *AzureSubscriber) HandleError(ctx context.Context, e FileReady, handlerError error) {
	logger.Error("failed to handle event", "event", e, "error", handlerError.Error())
	resp, err := as.Client.RejectEvents(ctx, []string{e.Event.LockToken}, nil)
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

func (as *AzureSubscriber) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Azure Event Subscriber"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE

	if as.Client == nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "Azure event subscriber not configured"
		return rsp
	}

	// Check via management API
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = "Failed to authenticate to Azure"
	}
	_, err = armeventgrid.NewClientFactory(as.Config.Subscription, cred, nil)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = fmt.Sprintf("Failed to connect to namespace %s", as.Config.Endpoint)
	}

	return rsp
}
