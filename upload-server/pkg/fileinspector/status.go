package fileinspector

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
)

type FileSystemUploadStatusInspector struct {
	BaseDir string
	ReportsDir string
}

func (fsusi *FileSystemUploadStatusInspector) InspectFileDeliveryStatus(_ context.Context, id string) ([]info.FileDeliveryStatus, error) {
	deliveries := []info.FileDeliveryStatus{}
	deliveryReportFilename := filepath.Join(fsusi.ReportsDir, id + event.TypeSeparator + reports.StageFileCopy)
	f, err := os.Open(deliveryReportFilename)
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

		issues := []string{}
		for _, issue := range report.StageInfo.Issues {
			issues = append(issues, issue.String())
		}

		deliveries = append(deliveries, info.FileDeliveryStatus{
			Status: report.StageInfo.Status,
			Name: content.DestinationName,
			Location: content.FileDestinationBlobUrl,
			DeliveredAt: report.StageInfo.EndProcessTime,
			Issues: issues,
		})
	}

	return deliveries, nil
}

func (fsusi *FileSystemUploadStatusInspector) InspectFileUploadStatus(ctx context.Context, id string) (string, error) {
	uploadStatusReportFilename := filepath.Join(fsusi.ReportsDir, id + event.TypeSeparator + reports.StageUploadStatus)
	uploadCompletedReportFilename := filepath.Join(fsusi.ReportsDir, id + event.TypeSeparator + reports.StageUploadCompleted)

	// check if the upload-complete file exists
	_, err := os.Stat(uploadCompletedReportFilename)
	if err == nil {
		// if the upload-complete file exists then the file is finished uploading
		return info.UploadComplete, nil
	} 

	// if the error from checking the upload-complete file is something other than
	// the file not existing, return an error
	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	// The status is now either not started or in progress 
	// we need to check the last line of the upload-status file
	// to check the offset and size
	f, err := os.Open(uploadStatusReportFilename)
	defer f.Close()

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", errors.Join(err, info.ErrNotFound)
		}
		return "", err
	}

	// Get the last line of the upload-status
	scanner := bufio.NewScanner(f)
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}
	
	// unmarshal the last line as a report
	var report reports.Report
	err = json.Unmarshal([]byte(lastLine), &report)
	if err != nil {
		return "", err
	}

	// check the status in the report.StageInfo
	if (report.StageInfo.Status == reports.StatusFailed) {
		return info.UploadFailed, nil
	}

	// retrieve the content from the report to get the file size and offset
	// report.Content is an map[string]any so we need to 
	// marshal it so we can unmarshal it into an UploadStatusContent
	// so that we can use it easier
	var content reports.UploadStatusContent
	b, err := json.Marshal(report.Content)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(b, &content)
	if err != nil {
		return "", err
	}

	// get the offset and size value so they can be compared
	offset := content.Offset
	size := content.Size

	if (offset == 0 && size == 0) {
		// if they are both 0 then the file hasn't started uploading
		return info.UploadNotStarted, nil
	}
	if (offset > 0 && size == 0) {
		// the size is not set until the file upload completes
		// so if the offset is greater than 0 but the size is 
		// still 0, then it is still being uploaded
		return info.UploadInProgress, nil
	}
	if (offset > 0 && size > 0 && offset == size) {
		// if the offset and size are greater than 0 and they are equal
		// then the file has completed upload
		// we shouldn't reach this state because we already checked
		// for file completion at the beginning, but there is a chance
		// that the file finished after we checked if the upload-completed file
		// existed but before we opened the upload-status file
		return info.UploadComplete, nil
	}

	// shouldn't get here, and if we have there's an issue
	return "", errors.New("Error determining the status of the file upload")
}

