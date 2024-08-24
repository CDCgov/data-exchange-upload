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

type FileSystemUploadInspector struct {
	BaseDir   string
	TusPrefix string
	ReportsDir string
}

func NewFileSystemUploadInspector(baseDir string, tusPrefix string, reportsDir string) *FileSystemUploadInspector {
	return &FileSystemUploadInspector{
		BaseDir:   baseDir,
		TusPrefix: tusPrefix,
		ReportsDir: reportsDir,
	}
}

func (fsui *FileSystemUploadInspector) InspectInfoFile(c context.Context, id string) (map[string]any, error) {
	// First, read in the .info file.
	infoFilename := filepath.Join(fsui.BaseDir, fsui.TusPrefix, id+".info")
	fileBytes, err := os.ReadFile(infoFilename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.Join(err, info.ErrNotFound)
		}
		return nil, err
	}

	// Deserialize to hash map.
	jsonMap := &info.InfoFileData{}
	if err := json.Unmarshal(fileBytes, jsonMap); err != nil {
		return nil, err
	}

	return jsonMap.MetaData, nil
}

func (fsui *FileSystemUploadInspector) InspectUploadedFile(c context.Context, id string) (map[string]any, error) {
	filename := filepath.Join(fsui.BaseDir, fsui.TusPrefix, id)
	fi, err := os.Stat(filename)
	if err != nil {
		return nil, errors.Join(err, info.ErrNotFound)
	}
	uploadedFileInfo := map[string]any{
		"updated_at": fi.ModTime(),
		"size_bytes": fi.Size(),
	}
	return uploadedFileInfo, nil
}

func (fsui *FileSystemUploadInspector) InspectFileStatus(_ context.Context, id string) (*info.FileStatus, error) {
	status := &info.FileStatus{
		Destinations: []info.FileDeliveryStatus{},
	}
	deliveryReportFilename := filepath.Join(fsui.ReportsDir, id + event.TypeSeparator + reports.StageFileCopy)
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

		status.Destinations = append(status.Destinations, info.FileDeliveryStatus{
			Status: report.StageInfo.Status,
			Name: "", // TODO need to store target in report
			// TODO need to get the content in a better way.  Maybe with generics?
			//Location: report.Content.(reports.FileCopyContent).FileDestinationBlobUrl, // TODO type check
		})
	}

	return status, nil
}
