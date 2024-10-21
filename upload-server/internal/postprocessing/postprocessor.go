package postprocessing

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
)

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

type PostProcessor struct {
	UploadBaseDir string
	UploadDir     string
}

type DeliveryMethod func(context.Context, string, delivery.Source, delivery.Destination) error

func ParallelStreamDelivery(ctx context.Context, id string, s delivery.Source, d delivery.Destination) error {
	dw, ok := d.(delivery.Writable)
	srcr, sok := s.(delivery.ReadInto)
	//TODO this isn't sustainable, implement a pattern registry or something
	if !sok || !ok {
		return errors.New("Cannot use this delivery method")
	}
	metadata, err := s.GetMetadata(ctx, id)
	if err != nil {
		return err
	}
	w, err := dw.Writer(ctx, id, metadata)
	if err != nil {
		return err
	}
	if err := srcr.ReadInto(ctx, id, w); err != nil {
		return err
	}
	return nil
}

var DeliveryMethods = []DeliveryMethod{
	ParallelStreamDelivery,
	delivery.Deliver,
}

func ProcessFileReadyEvent(ctx context.Context, e *event.FileReady) error {

	rb := reports.NewBuilder[reports.FileCopyContent](
		"1.0.0",
		reports.StageFileCopy,
		e.UploadId,
		reports.DispositionTypeAdd).SetStartTime(time.Now().UTC())

	defer func() {
		rb.SetEndTime(time.Now().UTC())
		report := rb.Build()
		reports.Publish(ctx, report)
	}()

	src, ok := delivery.GetSource("upload")
	if !ok {
		err := fmt.Errorf("failed to get source for file delivery %+v", e)
		rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
			Level:   reports.IssueLevelError,
			Message: err.Error(),
		})
		return err
	}
	dataStreamId := metadata.GetDataStreamID(e.Metadata)
	dataStreamRoute := metadata.GetDataStreamRoute(e.Metadata)
	d, ok := delivery.GetDestinationTarget(dataStreamId, dataStreamRoute, e.DestinationTarget)
	if !ok {
		err := fmt.Errorf("failed to get destination for file delivery %+v", e)
		rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
			Level:   reports.IssueLevelError,
			Message: err.Error(),
		})
		return err
	}
	//TODO get uri from the deliverer
	var deliveryErr error
	for _, method := range DeliveryMethods {
		deliveryErr = method(ctx, e.UploadId, src, d)
		if deliveryErr != nil {
			slog.Warn("Delivery Method failed", "method", method, "id", e.UploadId)
		}
	}
	if deliveryErr != nil {
		rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
			Level:   reports.IssueLevelError,
			Message: deliveryErr.Error(),
		})
		return deliveryErr
	}

	uri, err := d.URI(ctx, e.UploadId, e.Metadata)
	if err != nil {
		slog.Warn("failed to get uri for delivery", "event", e)
	}

	m, err := src.GetMetadata(ctx, e.UploadId)
	if err != nil {
		slog.Warn("failed to get metadata for report", "event", e)
	}
	rb.SetManifest(m)

	rb.SetContent(reports.FileCopyContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageFileCopy,
		},
		FileSourceBlobUrl:      e.SrcUrl,
		FileDestinationBlobUrl: uri,
		DestinationName:        e.DestinationTarget,
	})

	return err
}
