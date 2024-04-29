package fileinspector

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
)

type FileSystemUploadInspector struct {
	BaseDir   string
	TusPrefix string
}

func NewFileSystemUploadInspector(baseDir string, tusPrefix string) *FileSystemUploadInspector {
	return &FileSystemUploadInspector{
		BaseDir:   baseDir,
		TusPrefix: tusPrefix,
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
