package postprocessing

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"

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

func DeliverFileReadyEvent(ctx context.Context, e *event.FileReady) error {

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
