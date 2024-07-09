package upload

import (
	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

func ReportUploadStatus(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	// Get values from event
	uploadId := event.Upload.ID
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	uploadMetadata := event.Upload.MetaData

	rcb := reports.NewReportContentBuilder[reports.UploadStatusContent]().SetContent(reports.UploadStatusContent{
		ReportContent: reports.ReportContent{
			SchemaVersion: "1.0.0",
			SchemaName:    reports.StageUploadStatus,
		},
		Filename: metadataPkg.GetFilename(uploadMetadata),
		Tguid:    uploadId,
		Offset:   strconv.FormatInt(uploadOffset, 10),
		Size:     strconv.FormatInt(uploadSize, 10),
	})

	report := reports.NewBuilder(
		"1.0.0",
		reports.StageUploadStatus,
		uploadId,
		uploadMetadata,
		reports.DispositionTypeReplace,
		rcb).Build()

	logger.Info("REPORT", "report", report)
	reports.Publish(event.Context, report)

	return resp, nil
}

func ReportUploadStarted(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	uploadId := event.Upload.ID
	manifest := event.Upload.MetaData
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	logger.Info("Attempting to report upload started")

	uploadStartedBuilder := reports.NewReportContentBuilder[reports.UploadLifecycleContent]().SetContent(reports.UploadLifecycleContent{
		ReportContent: reports.ReportContent{
			SchemaVersion: "1.0.0",
			SchemaName:    reports.StageUploadStarted,
		},
		Status: "success",
	})
	report := reports.NewBuilder(
		"1.0.0",
		reports.StageUploadStarted,
		uploadId,
		manifest,
		reports.DispositionTypeAdd,
		uploadStartedBuilder).Build()
	reports.Publish(event.Context, report)

	uploadStatusBuilder := reports.NewReportContentBuilder[reports.UploadStatusContent]().SetContent(reports.UploadStatusContent{
		ReportContent: reports.ReportContent{
			SchemaVersion: "1.0.0",
			SchemaName:    reports.StageUploadStatus,
		},
		Filename: metadataPkg.GetFilename(manifest),
		Tguid:    uploadId,
		Offset:   strconv.FormatInt(uploadOffset, 10),
		Size:     strconv.FormatInt(uploadSize, 10),
	})

	report = reports.NewBuilder(
		"1.0.0",
		reports.StageUploadStatus,
		uploadId,
		manifest,
		reports.DispositionTypeReplace,
		uploadStatusBuilder).SetStartTime(time.Now().UTC()).Build()

	logger.Info("REPORT upload-status", "report", report)
	reports.Publish(event.Context, report)

	return resp, nil
}

func ReportUploadComplete(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	uploadId := event.Upload.ID
	manifest := event.Upload.MetaData
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	logger.Info("Attempting to report upload completed", "uploadId", uploadId)

	uploadCompletedBuilder := reports.NewReportContentBuilder[reports.UploadLifecycleContent]().SetContent(reports.UploadLifecycleContent{
		ReportContent: reports.ReportContent{
			SchemaVersion: "1.0.0",
			SchemaName:    reports.StageUploadCompleted,
		},
		Status: "success",
	})
	report := reports.NewBuilder(
		"1.0.0",
		reports.StageUploadCompleted,
		uploadId,
		manifest,
		reports.DispositionTypeAdd,
		uploadCompletedBuilder).Build()
	reports.Publish(event.Context, report)

	rcb := reports.NewReportContentBuilder[reports.UploadStatusContent]().SetContent(reports.UploadStatusContent{
		ReportContent: reports.ReportContent{
			SchemaVersion: "1.0.0",
			SchemaName:    reports.StageUploadStatus,
		},
		Filename: metadataPkg.GetFilename(manifest),
		Tguid:    uploadId,
		Offset:   strconv.FormatInt(uploadOffset, 10),
		Size:     strconv.FormatInt(uploadSize, 10),
	})

	report = reports.NewBuilder(
		"1.0.0",
		reports.StageUploadStatus,
		uploadId,
		manifest,
		reports.DispositionTypeReplace,
		rcb).SetEndTime(time.Now().UTC()).SetStatus("success").Build()

	logger.Info("REPORT upload-status", "report", report)
	reports.Publish(event.Context, report)

	return resp, nil
}
