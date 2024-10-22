package event

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus/admin"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
)

var logger *slog.Logger
var FileReadyPublisher Publisher[*FileReady]

const TypeSeparator = "_"

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
	p = &MemoryBus[T]{
		Dir:  appConfig.LocalEventsFolder,
		Chan: make(chan T),
	}

	if appConfig.PublisherConnection != nil {
		var err error
		p, err = NewAzurePublisher[T](ctx, *appConfig.PublisherConnection)
		health.Register(p)
		return p, err
	}

	return p, nil
}

type MemoryBus[T Identifiable] struct {
	Dir    string
	Chan   chan T
	closed bool
}

func (mp *MemoryBus[T]) Publish(ctx context.Context, event T) error {
	err := os.MkdirAll(mp.Dir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	filename := filepath.Join(mp.Dir, event.Identifier()+TypeSeparator+event.Type())
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

	if mp.Chan != nil && !mp.closed {
		go func() {
			mp.Chan <- event
		}()
	}
	return nil
}

func (mp *MemoryBus[T]) Close() error {
	mp.closed = true
	if !mp.closed && mp.Chan != nil {
		close(mp.Chan)
	}
	return nil
}

func (mp *MemoryBus[T]) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "Memory Publisher"
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

func (ms *MemoryBus[T]) GetBatch(ctx context.Context, _ int) ([]T, error) {
	select {
	case <-ctx.Done():
		return nil, nil
	case evt := <-ms.Chan:
		return []T{evt}, nil
	}
}

func (ms *MemoryBus[T]) HandleSuccess(_ context.Context, e T) error {
	logger.Info("successfully handled event", "event", e)
	return nil
}

func (ms *MemoryBus[T]) HandleError(ctx context.Context, e T, err error) error {
	logger.Error("failed to handle event", "event", e, "error", err.Error())
	if e.RetryCount() < MaxRetries {
		e.IncrementRetryCount()
		// Retrying in a separate go routine so this doesn't block on channel write.
		go func() {
			if ctx.Err() == nil {
				ms.Chan <- e
			}
		}()
	}
	return nil
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
