package postprocessing

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
)

type PostProcessor struct {
	UploadBaseDir string
	UploadDir     string
}

func ProcessFileReadyEvent(ctx context.Context, e *event.FileReady) error {

	slog.Info("starting file copy", "uploadId", e.UploadId)

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
	uri, err := delivery.Deliver(ctx, e.UploadId, e.Path, src, d)

	if err != nil {
		rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
			Level:   reports.IssueLevelError,
			Message: err.Error(),
		})
		return err
	}
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
