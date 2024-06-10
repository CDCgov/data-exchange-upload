package upload

import (
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
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

	content := &models.UploadStatusContent{
		ReportContent: models.ReportContent{
			SchemaVersion: "1.0",
			SchemaName:    "upload",
		},
		Filename: metadata.GetFilename(uploadMetadata),
		Tguid:    uploadId,
		Offset:   strconv.FormatInt(uploadOffset, 10),
		Size:     strconv.FormatInt(uploadSize, 10),
	}

	report := &models.Report{
		UploadID:        uploadId,
		DataStreamID:    metadata.GetDataStreamID(uploadMetadata),
		DataStreamRoute: metadata.GetDataStreamRoute(uploadMetadata),
		StageName:       "dex-upload-status",
		ContentType:     "json",
		DispositionType: "replace",
		Content:         content,
	}

	logger.Info("REPORT", "report", report)
	reports.Publish(event.Context, report)

	return resp, nil
}

func ReportUploadStarted(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	uploadId := event.Upload.ID
	manifest := event.Upload.MetaData
	logger.Info("Attempting to report upload started")
	content := &models.UploadLifecycleContent{
		ReportContent: models.ReportContent{
			SchemaVersion: "1.0",
			SchemaName:    "dex-upload-started",
		},
		Status: "success",
	}

	report := &models.Report{
		UploadID:        uploadId,
		DataStreamID:    metadata.GetDataStreamID(manifest),
		DataStreamRoute: metadata.GetDataStreamRoute(manifest),
		StageName:       "dex-upload-started",
		ContentType:     "json",
		DispositionType: "add",
		Content:         content,
	}

	logger.Info("REPORT upload-started", "report", report)
	reports.Publish(event.Context, report)
	return resp, nil
}

func ReportUploadComplete(event handler.HookEvent, resp hooks.HookResponse) (hooks.HookResponse, error) {
	uploadId := event.Upload.ID
	manifest := event.Upload.MetaData
	logger.Info("Attempting to report upload completed", "uploadId", uploadId)
	content := &models.UploadLifecycleContent{
		ReportContent: models.ReportContent{
			SchemaVersion: "1.0",
			SchemaName:    "dex-upload-complete",
		},
		Status: "success",
	}

	report := &models.Report{
		UploadID:        uploadId,
		DataStreamID:    metadata.GetDataStreamID(manifest),
		DataStreamRoute: metadata.GetDataStreamRoute(manifest),
		StageName:       "dex-upload-complete",
		ContentType:     "json",
		DispositionType: "add",
		Content:         content,
	}

	logger.Info("REPORT upload-completed", "report", report)
	reports.Publish(event.Context, report)
	return resp, nil
}
