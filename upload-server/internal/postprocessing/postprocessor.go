package postprocessing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/prometheus/client_golang/prometheus"
)

type PostProcessor struct {
	UploadBaseDir string
	UploadDir     string
}

func ProcessFileReadyEvent(ctx context.Context, e *event.FileReady) error {
	if e == nil || e.UploadId == "" {
		return fmt.Errorf("malformed file ready event %+v", e)
	}
	slog.Info("starting file copy", "uploadId", e.UploadId)
	metrics.EventsCounter.With(prometheus.Labels{metrics.Labels.EventType: e.Type(), metrics.Labels.EventOp: "subscribe"}).Inc()

	rb := reports.NewBuilder[reports.FileCopyContent](
		"1.0.0",
		reports.StageFileCopy,
		e.UploadId,
		reports.DispositionTypeAdd).SetStartTime(time.Now().UTC()).SetContent(reports.FileCopyContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageFileCopy,
		},
		FileSourceBlobUrl: e.SrcUrl,
		DestinationName:   e.DestinationTarget,
	})

	defer func() {
		rb.SetEndTime(time.Now().UTC())
		report := rb.Build()
		slog.Info("REPORT blob-file-copy", "report", report, "uploadId", e.UploadId)
		reports.Publish(ctx, report)

		slog.Info("file-copy report complete", "uploadId", e.UploadId)
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
	d, ok := delivery.GetTarget(e.DestinationTarget)
	if !ok {
		err := fmt.Errorf("failed to get destination for file delivery %+v", e)
		rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
			Level:   reports.IssueLevelError,
			Message: err.Error(),
		})
		return err
	}

	metrics.ActiveDeliveries.With(prometheus.Labels{"target": e.DestinationTarget}).Inc()
	metrics.DeliveryTotals.With(prometheus.Labels{"target": e.DestinationTarget, "result": "started"}).Inc()
	uri, err := delivery.Deliver(ctx, e.UploadId, e.Path, src, d)
	metrics.ActiveDeliveries.With(prometheus.Labels{"target": e.DestinationTarget}).Dec()

	if err != nil {
		slog.Error("failed to deliver file", "target", uri, "error", err)
		rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
			Level:   reports.IssueLevelError,
			Message: err.Error(),
		})
		metrics.DeliveryTotals.With(prometheus.Labels{"target": e.DestinationTarget, "result": "failed"}).Inc()
		return err
	}
	metrics.DeliveryTotals.With(prometheus.Labels{"target": e.DestinationTarget, "result": "completed"}).Inc()
	slog.Info("file delivered", "event", e, "uploadId", e.UploadId) // Is this necessary?

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
