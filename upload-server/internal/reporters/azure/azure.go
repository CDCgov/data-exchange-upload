package reporters

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
//	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
//	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
//	"github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters"
//)
//
//type ServiceBusReporter struct {
//	Context     context.Context
//	Sender      *azservicebus.Sender
//	AdminClient *admin.Client
//	QueueName   string
//}
//
//func (sb *ServiceBusReporter) Publish(ctx context.Context, r reporters.Identifiable) error {
//	b, err := json.Marshal(r)
//	if err != nil {
//		return err
//	}
//
//	m := &azservicebus.Message{
//		Body: b,
//	}
//
//	return sb.Sender.SendMessage(ctx, m, nil)
//}
//
//func (sb *ServiceBusReporter) Close() error {
//	return sb.Sender.Close(sb.Context)
//}
//
//func (sb *ServiceBusReporter) Health(ctx context.Context) models.ServiceHealthResp {
//	var shr models.ServiceHealthResp
//	shr.Service = "Reports"
//
//	// Get the service bus queue.
//	queueResp, err := sb.AdminClient.GetQueue(ctx, sb.QueueName, nil)
//	if err != nil {
//		return shr.BuildErrorResponse(err)
//	}
//
//	// Check the queue status is active.
//	if *queueResp.Status != admin.EntityStatusActive {
//		return shr.BuildErrorResponse(fmt.Errorf("service bus queue status: %s", *queueResp.Status))
//	}
//
//	// all good
//	shr.Status = models.STATUS_UP
//	shr.HealthIssue = models.HEALTH_ISSUE_NONE
//	return shr
//}
