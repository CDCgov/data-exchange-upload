package sbhealth

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
)

// Service Bus
type ServiceBusHealth struct {
	Client *admin.Client
	Queue  string
}

func New(appConfig appconfig.AppConfig) (*ServiceBusHealth, error) {
	// TODO set retry options.
	client, err := admin.NewClientFromConnectionString(appConfig.ServiceBusConnectionString, nil)
	if err != nil {
		return nil, err
	}

	return &ServiceBusHealth{
		Client: client,
		Queue:  appConfig.ReportQueueName,
	}, nil
}

func (sbHealth ServiceBusHealth) Health(ctx context.Context) models.ServiceHealthResp {
	var shr models.ServiceHealthResp
	shr.Service = models.SERVICE_BUS

	// Get the service bus queue.
	queueResp, err := sbHealth.Client.GetQueue(ctx, sbHealth.Queue, nil)
	if err != nil {
		return serviceBusDown(err)
	}

	// Check the queue status is active.
	if *queueResp.Status != admin.EntityStatusActive {
		return serviceBusDown(fmt.Errorf("service bus queue status: %s", *queueResp.Status))
	}

	// all good
	shr.Status = models.STATUS_UP
	shr.HealthIssue = models.HEALTH_ISSUE_NONE
	return shr
}

func serviceBusDown(err error) models.ServiceHealthResp {
	return models.ServiceHealthResp{
		Service:     models.SERVICE_BUS,
		Status:      models.STATUS_DOWN,
		HealthIssue: err.Error(),
	}
}
