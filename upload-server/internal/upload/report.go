package upload

import (
	"time"

	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func ReportUploadStatus(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	// Get values from event
	uploadId := event.Upload.ID
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	uploadMetadata := event.Upload.MetaData
	logger := sloger.FromContext(event.Context)

	logger.Info("starting upload-status report")

	report := reports.NewBuilderWithManifest[reports.UploadStatusContent](
		"1.0.0",
		reports.StageUploadStatus,
		uploadId,
		uploadMetadata,
		reports.DispositionTypeReplace).SetContent(reports.UploadStatusContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageUploadStatus,
		},
		Filename: metadataPkg.GetFilename(uploadMetadata),
		Tguid:    uploadId,
		Offset:   uploadOffset,
		Size:     uploadSize,
	}).Build()

	logger.Info("REPORT upload-status", "report", report)
	reports.Publish(event.Context, report)

	logger.Info("upload-status report complete")

	return resp, nil
}

func ReportUploadStarted(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	uploadId := event.Upload.ID
	manifest := event.Upload.MetaData
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	logger := sloger.FromContext(event.Context)

	logger.Info("starting upload-started report")

	report := reports.NewBuilderWithManifest[reports.UploadLifecycleContent](
		"1.0.0",
		reports.StageUploadStarted,
		uploadId,
		manifest,
		reports.DispositionTypeAdd).SetContent(reports.UploadLifecycleContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageUploadStarted,
		},
		Status: reports.StatusSuccess,
	}).Build()

	logger.Info("REPORT upload-started", "report", report)
	reports.Publish(event.Context, report)

	report = reports.NewBuilderWithManifest[reports.UploadStatusContent](
		"1.0.0",
		reports.StageUploadStatus,
		uploadId,
		manifest,
		reports.DispositionTypeReplace).SetStartTime(time.Now().UTC()).SetContent(reports.UploadStatusContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageUploadStatus,
		},
		Filename: metadataPkg.GetFilename(manifest),
		Tguid:    uploadId,
		Offset:   uploadOffset,
		Size:     uploadSize,
	}).Build()

	logger.Info("REPORT upload-status", "report", report)
	reports.Publish(event.Context, report)

	logger.Info("upload-started report complete")

	return resp, nil
}

func ReportUploadComplete(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	uploadId := event.Upload.ID
	manifest := event.Upload.MetaData
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	logger := sloger.FromContext(event.Context)

	logger.Info("starting upload-completed report")

	report := reports.NewBuilderWithManifest[reports.UploadLifecycleContent](
		"1.0.0",
		reports.StageUploadCompleted,
		uploadId,
		manifest,
		reports.DispositionTypeAdd).SetContent(reports.UploadLifecycleContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageUploadCompleted,
		},
		Status: reports.StatusSuccess,
	}).Build()

	logger.Info("REPORT upload-completed", "report", report)
	reports.Publish(event.Context, report)

	report = reports.NewBuilderWithManifest[reports.UploadStatusContent](
		"1.0.0",
		reports.StageUploadStatus,
		uploadId,
		manifest,
		reports.DispositionTypeReplace).SetEndTime(time.Now().UTC()).SetContent(reports.UploadStatusContent{
		ReportContent: reports.ReportContent{
			ContentSchemaVersion: "1.0.0",
			ContentSchemaName:    reports.StageUploadStatus,
		},
		Filename: metadataPkg.GetFilename(manifest),
		Tguid:    uploadId,
		Offset:   uploadOffset,
		Size:     uploadSize,
	}).Build()

	logger.Info("REPORT upload-status", "report", report)
	reports.Publish(event.Context, report)

	logger.Info("upload-completed report complete")
	return resp, nil
}
