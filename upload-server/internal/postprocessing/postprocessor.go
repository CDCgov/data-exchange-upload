package postprocessing

import (
	"context"
	"fmt"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
)

type PostProcessor struct {
	UploadBaseDir string
	UploadDir     string
}

func ProcessFileReadyEvent(ctx context.Context, e *event.FileReady) error {
	logger := sloger.FromContext(ctx)

	if e == nil || e.UploadId == "" {
		return fmt.Errorf("malformed file ready event %+v", e)
	}

	logger.Info("starting file copy")

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
		logger.Info("REPORT blob-file-copy", "report", report)
		reports.Publish(ctx, report)

		logger.Info("file-copy report complete")
	}()

	src, ok := delivery.GetSource(delivery.UploadSrc)
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
		logger.Error("failed to deliver file", "target", uri, "error", err)
		rb.SetStatus(reports.StatusFailed).AppendIssue(reports.ReportIssue{
			Level:   reports.IssueLevelError,
			Message: err.Error(),
		})
		return err
	}
	logger.Info("file delivered", "event", e) // Is this necessary?

	m, err := src.GetMetadata(ctx, e.UploadId)
	if err != nil {
		logger.Warn("failed to get metadata for report", "event", e)
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
