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
	"io"
	"log/slog"
	"os"
	"reflect"
	"strings"
)

var logger *slog.Logger
var FileReadyPublisher Publisher[*FileReady]

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

func InitFileReadyPublisher(ctx context.Context, appConfig appconfig.AppConfig) error {
	p, err := NewEventPublisher[*FileReady](ctx, appConfig)
	FileReadyPublisher = p
	return err
}

type Publisher[T Identifiable] interface {
	health.Checkable
	io.Closer
	Publish(ctx context.Context, event T) error
}

func NewEventPublisher[T Identifiable](ctx context.Context, appConfig appconfig.AppConfig) (Publisher[T], error) {
	var p Publisher[T]
	c, err := GetChannel[T]()
	if err != nil {
		return nil, err
	}

	p = &MemoryPublisher[T]{
		Dir:  appConfig.LocalEventsFolder,
		Chan: c,
	}

	if appConfig.PublisherConnection != nil {
		p, err = NewAzurePublisher[T](ctx, *appConfig.PublisherConnection)
		health.Register(p)
	}

	return p, nil
}

type MemoryPublisher[T Identifiable] struct {
	Dir  string
	Chan chan T
}

func (mp *MemoryPublisher[T]) Publish(_ context.Context, event T) error {
	err := os.Mkdir(mp.Dir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	filename := mp.Dir + "/" + event.Identifier()
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

	if mp.Chan != nil {
		mp.Chan <- event
	}
	return nil
}

func (mp *MemoryPublisher[T]) Close() error {
	return nil
}

func (mp *MemoryPublisher[T]) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Publisher"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

type AzurePublisher[T Identifiable] struct {
	Context     context.Context
	Sender      *azservicebus.Sender
	Config      appconfig.AzureQueueConfig
	AdminClient *admin.Client
}

func (ap *AzurePublisher[T]) Publish(ctx context.Context, event T) error {
	b, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ap.Sender.SendMessage(ctx, &azservicebus.Message{
		Body: b,
	}, nil)
}

func (ap *AzurePublisher[T]) Close() error {
	return ap.Sender.Close(ap.Context)
}

func (ap *AzurePublisher[T]) Health(ctx context.Context) (rsp models.ServiceHealthResp) {
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE

	if ap.Config.Queue != "" {
		rsp.Service = fmt.Sprintf("Event Publishing %s", ap.Config.Queue)
		queueResp, err := ap.AdminClient.GetQueue(ctx, ap.Config.Queue, nil)
		if err != nil {
			return rsp.BuildErrorResponse(err)
		}
		if queueResp == nil {
			return rsp.BuildErrorResponse(fmt.Errorf("nil queue response"))
		}
		if *queueResp.Status != admin.EntityStatusActive {
			return rsp.BuildErrorResponse(fmt.Errorf("service bus queue %s status: %s", ap.Config.Queue, *queueResp.Status))
		}
	}

	if ap.Config.Topic != "" {
		rsp.Service = fmt.Sprintf("Event Publishing %s", ap.Config.Topic)
		topicResp, err := ap.AdminClient.GetTopic(ctx, ap.Config.Topic, nil)
		if err != nil {
			return rsp.BuildErrorResponse(err)
		}
		if *topicResp.Status != admin.EntityStatusActive {
			return rsp.BuildErrorResponse(fmt.Errorf("service bus topic %s status: %s", ap.Config.Topic, *topicResp.Status))
		}
	}

	return rsp
}
