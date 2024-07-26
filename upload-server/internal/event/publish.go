package event

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"log/slog"
	"os"
	"reflect"
	"strings"
)

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

type Publisher interface {
	health.Checkable
	Publish(ctx context.Context, event FileReady) error
}

type MemoryPublisher struct {
	Dir string
}

type AzurePublisher struct {
	EventType   string
	Sender      *azservicebus.Sender
	Config      appconfig.AzureQueueConfig
	AdminClient *admin.Client
}

func (mp *MemoryPublisher) Publish(_ context.Context, event FileReady) error {
	err := os.Mkdir(mp.Dir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	filename := mp.Dir + "/" + event.UploadId
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// write event to file.
	encoder := json.NewEncoder(f)
	err = encoder.Encode(event)
	if err != nil {
		return err
	}

	fileReadyChan <- event
	return nil
}

func (mp *MemoryPublisher) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Publisher"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (ap *AzurePublisher) Publish(ctx context.Context, event FileReady) error {
	//evt, err := messaging.NewCloudEvent("upload", FileReadyEventType, event, nil)
	//if err != nil {
	//	return err
	//}

	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ap.Sender.SendMessage(ctx, &azservicebus.Message{
		Body: b,
	}, nil)

	//_, err = ap.Client.SendEvent(ctx, &evt, nil)
	//return err
}

func (ap *AzurePublisher) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = fmt.Sprintf("%s Event Publishing", ap.EventType)
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE

	topicResp, err := ap.AdminClient.GetTopic(ctx, ap.Config.Topic, nil)
	if err != nil {
		return rsp.BuildErrorResponse(err)
	}

	if *topicResp.Status != admin.EntityStatusActive {
		return rsp.BuildErrorResponse(fmt.Errorf("service bus topic %s status: %s", ap.Config.Topic, *topicResp.Status))
	}

	//if ap.Client == nil {
	//	rsp.Status = models.STATUS_DOWN
	//	rsp.HealthIssue = "Azure event publisher not configured"
	//	return rsp
	//}
	//
	//// Check via management API
	//cred, err := azidentity.NewDefaultAzureCredential(nil)
	//if err != nil {
	//	rsp.Status = models.STATUS_DOWN
	//	rsp.HealthIssue = "Failed to authenticate to Azure"
	//}
	//_, err = armeventgrid.NewClientFactory(ap.Config.Subscription, cred, nil)
	//if err != nil {
	//	rsp.Status = models.STATUS_DOWN
	//	rsp.HealthIssue = fmt.Sprintf("Failed to connect to namespace %s", ap.Config.Endpoint)
	//}

	return rsp
}
