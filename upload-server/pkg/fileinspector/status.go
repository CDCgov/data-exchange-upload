package fileinspector

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/reports"
	"os"
	"path/filepath"
)

type FileSystemUploadStatusInspector struct {
	BaseDir string
	ReportsDir string
}

func (fsusi *FileSystemUploadStatusInspector) InspectFileStatus(_ context.Context, id string) (*info.DeliveryStatus, error) {
	status := &info.DeliveryStatus{
		Destinations: []info.FileDeliveryStatus{},
	}
	deliveryReportFilename := filepath.Join(fsusi.ReportsDir, id + event.TypeSeparator + reports.StageFileCopy)
	f, err := os.Open(deliveryReportFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.Join(err, info.ErrNotFound)
		}
		return status, err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		l := scanner.Text()

		var report reports.Report
		err := json.Unmarshal([]byte(l), &report)
		if err != nil {
			return status, err
		}

		var content reports.FileCopyContent
		b, err := json.Marshal(report.Content)
		if err != nil {
			return status, err
		}
		err = json.Unmarshal(b, &content)
		if err != nil {
			return status, err
		}

		issues := []string{}
		for _, issue := range report.StageInfo.Issues {
			issues = append(issues, issue.String())
		}

		status.Destinations = append(status.Destinations, info.FileDeliveryStatus{
			Status: report.StageInfo.Status,
			Name: content.DestinationName,
			Location: content.FileDestinationBlobUrl,
			DeliveredAt: report.StageInfo.EndProcessTime,
			Issues: issues,
		})
	}

	return status, nil
}