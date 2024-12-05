package upload

import (
	"log/slog"
	"reflect"
	"strings"
	"time"

	metadataPkg "github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

func TrimS3MultipartID(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	event.Upload.ID, _, _ = strings.Cut(event.Upload.ID, "+")
	return resp, nil
}

func ReportUploadStatus(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	// Get values from event
	uploadId := event.Upload.ID
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	uploadMetadata := event.Upload.MetaData

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

	logger.Info("REPORT", "report", report)
	reports.Publish(event.Context, report)

	return resp, nil
}

func ReportUploadStarted(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	uploadId := event.Upload.ID
	manifest := event.Upload.MetaData
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	logger.Info("Attempting to report upload started")

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

	return resp, nil
}

func ReportUploadComplete(event *handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	uploadId := event.Upload.ID
	manifest := event.Upload.MetaData
	uploadOffset := event.Upload.Offset
	uploadSize := event.Upload.Size
	logger.Info("Attempting to report upload completed", "uploadId", uploadId)

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

	return resp, nil
}
