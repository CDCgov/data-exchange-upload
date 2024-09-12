package fileinspector

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
)

type FileSystemUploadStatusInspector struct {
	BaseDir    string
	ReportsDir string
}

func (fsusi *FileSystemUploadStatusInspector) InspectFileDeliveryStatus(_ context.Context, id string) ([]info.FileDeliveryStatus, error) {
	deliveries := []info.FileDeliveryStatus{}
	deliveryReportFilename := filepath.Join(fsusi.ReportsDir, id+event.TypeSeparator+reports.StageFileCopy)
	f, err := os.Open(deliveryReportFilename)
	defer f.Close()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.Join(err, info.ErrNotFound)
		}
		return deliveries, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()

		var report reports.Report
		err := json.Unmarshal([]byte(l), &report)
		if err != nil {
			return deliveries, err
		}

		var content reports.FileCopyContent
		b, err := json.Marshal(report.Content)
		if err != nil {
			return deliveries, err
		}
		err = json.Unmarshal(b, &content)
		if err != nil {
			return deliveries, err
		}

		deliveries = append(deliveries, info.FileDeliveryStatus{
			Status:      report.StageInfo.Status,
			Name:        content.DestinationName,
			Location:    content.FileDestinationBlobUrl,
			DeliveredAt: report.StageInfo.EndProcessTime,
			Issues:      report.StageInfo.Issues,
		})
	}

	return deliveries, nil
}

func (fsusi *FileSystemUploadStatusInspector) InspectFileUploadStatus(ctx context.Context, id string) (info.FileUploadStatus, error) {
	// check if the upload-completed file exists
	uploadCompletedReportFilename := filepath.Join(fsusi.ReportsDir, id + event.TypeSeparator + reports.StageUploadCompleted)
	uploadCompletedFileInfo, errComplete := os.Stat(uploadCompletedReportFilename)
	if errComplete == nil {
		lastChunkReceived := uploadCompletedFileInfo.ModTime().UTC().Format(time.RFC3339)
		// if the upload-completed file exists then the file is finished uploading
		return info.FileUploadStatus{
			Status: info.UploadComplete,
			LastChunkReceived: lastChunkReceived,
		}, nil
	} 

	// if the error from checking the upload-completed file is something other than
	// the file not existing, return the error
	if !errors.Is(errComplete, os.ErrNotExist) {
		return info.FileUploadStatus{}, errComplete
	}

	// get the file info for the upload-started file
	uploadStartedReportFilename := filepath.Join(fsusi.ReportsDir, id + event.TypeSeparator + reports.StageUploadStarted)
	uploadStartedFileInfo, errStart := os.Stat(uploadStartedReportFilename)
	if errStart != nil {
		if errors.Is(errStart, os.ErrNotExist) {
			return info.FileUploadStatus{}, errors.Join(errStart, info.ErrNotFound)
		}
		return info.FileUploadStatus{}, errStart
	}

	// get the file info for the upload-status file
	uploadStatusReportFilename := filepath.Join(fsusi.ReportsDir, id + event.TypeSeparator + reports.StageUploadStatus)
	uploadStatusReportFileInfo, errStatus := os.Stat(uploadStatusReportFilename)
	if errStatus != nil {
		if errors.Is(errStatus, os.ErrNotExist) {
			return info.FileUploadStatus{}, errors.Join(errStatus, info.ErrNotFound)
		}
		return info.FileUploadStatus{}, errStatus
	}

	uploadStartedModTime := uploadStartedFileInfo.ModTime()
	uploadStatusModTime := uploadStatusReportFileInfo.ModTime()

	lastChunkReceived := uploadStatusModTime.UTC().Format(time.RFC3339)

	if (uploadStartedModTime.Unix() == uploadStatusModTime.Unix()) {
		// when the file upload is initiated the upload-started and upload-status reports 
		// are created at the same time, so if the file modified times are still equal 
		// it means the file hasn't started uploading yet
		return info.FileUploadStatus{
			Status: info.UploadInitiated,
			LastChunkReceived: lastChunkReceived,
		}, nil
	}

	if (uploadStartedModTime.Unix() < uploadStatusModTime.Unix()) {
		// if the file modified time of the upload-status report is later than the
		// upload-started report, then the file is being uploaded
		return info.FileUploadStatus{
			Status: info.UploadInProgress,
			LastChunkReceived: lastChunkReceived,
		}, nil
	}

	return info.FileUploadStatus{}, errors.New("Unable to determine the status of the upload")
}
