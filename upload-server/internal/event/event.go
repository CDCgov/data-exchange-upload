package event

import (
	"context"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus"
)

const FileReadyEventType = "FileReady"

var FileReadyPublisher Publishers[*FileReady]

var MaxRetries = 5
var MaxMessages = 3

type Retryable interface {
	RetryCount() int
	IncrementRetryCount()
}

// TODO better name for this interface would be Subscribable or Queueable or similar
type Identifiable interface {
	Retryable
	Identifier() string
	GetUploadID() string
	Type() string
	SetIdentifier(id string)
	SetType(t string)
}

type Event struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	RetryCount int    `json:"retry_count"`
}

type FileReady struct {
	Event
	UploadId          string `json:"upload_id"`
	SrcUrl            string `json:"src_url"`
	Path              string `json:"path"`
	DestinationTarget string `json:"deliver_target"`
	Metadata          map[string]string
}

func (fr *FileReady) RetryCount() int {
	return fr.Event.RetryCount
}

func (fr *FileReady) IncrementRetryCount() {
	fr.Event.RetryCount++
}

func (fr *FileReady) Type() string {
	return fr.Event.Type
}

func (fr *FileReady) SetIdentifier(id string) {
	fr.ID = id
}

func (fr *FileReady) SetType(t string) {
	fr.Event.Type = t
}

func (fr *FileReady) Identifier() string {
	return fr.UploadId + fr.DestinationTarget
}

func (fr *FileReady) GetUploadID() string {
	return fr.UploadId
}

func NewFileReadyEvent(uploadId string, metadata map[string]string, path, target string) *FileReady {
	return &FileReady{
		Event: Event{
			Type: FileReadyEventType,
		},
		Path:              path,
		UploadId:          uploadId,
		Metadata:          metadata,
		DestinationTarget: target,
	}
}

func MetricsProcessor[T Identifiable](next func(context.Context, T) error) func(context.Context, T) error {
	return func(ctx context.Context, e T) error {
		var err error
		metrics.EventsCounter.With(prometheus.Labels{metrics.Labels.EventType: e.Type(), metrics.Labels.EventOp: "subscribe"}).Inc()

		fr, isFileReady := any(e).(*FileReady)
		if isFileReady {
			defer func() {
				metrics.ActiveDeliveries.With(prometheus.Labels{"target": fr.DestinationTarget}).Dec()
				if err != nil {
					metrics.DeliveryTotals.With(prometheus.Labels{"target": fr.DestinationTarget, "result": metrics.DeliveryResultFailed}).Inc()
				} else {
					metrics.DeliveryTotals.With(prometheus.Labels{"target": fr.DestinationTarget, "result": metrics.DeliveryResultCompleted}).Inc()
				}
			}()
			metrics.ActiveDeliveries.With(prometheus.Labels{"target": fr.DestinationTarget}).Inc()
			metrics.DeliveryTotals.With(prometheus.Labels{"target": fr.DestinationTarget, "result": metrics.DeliveryResultStarted}).Inc()
		}

		err = next(ctx, e)
		return err
	}
}

func UploadIDLoggerProcessor[T Identifiable](next func(context.Context, T) error) func(context.Context, T) error {
	return func(ctx context.Context, e T) error {
		c, _ := sloger.SetInContext(ctx, "uploadId", e.GetUploadID())
		return next(c, e)
	}
}
