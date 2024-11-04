package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/models"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type FileDestination struct {
	ToPath       string `yaml:"path"`
	Name         string `yaml:"name"`
	PathTemplate string `yaml:"path_template"`
}

func (fd *FileDestination) DestinationType() string {
	return storageTypeLocalFile
}

func (fd *FileDestination) Copy(ctx context.Context, id string, path string, source *Source,
	metadata map[string]string, _ int64, _ int) (string, error) {
	s := *source
	reader, _ := s.Reader(ctx, id)
	return fd.Upload(ctx, path, reader, metadata)
}

func (fd *FileDestination) Upload(_ context.Context, path string, r io.Reader, m map[string]string) (string, error) {
	loc := filepath.Join(fd.ToPath, path)
	if err := os.MkdirAll(filepath.Dir(loc), 0755); err != nil {
		return "", err
	}
	dest, err := os.Create(loc)
	if err != nil {
		return path, err
	}
	defer dest.Close()
	if _, err := io.Copy(dest, r); err != nil {
		return dest.Name(), err
	}
	return dest.Name(), nil
}

func (fd *FileDestination) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "File Delivery Target " + fd.Name
	info, err := os.Stat(fd.ToPath)
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if !info.IsDir() {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = fmt.Sprintf("%s is not a directory", fd.ToPath)
		return rsp
	}
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}

type FileSource struct {
	FS fs.FS
}

func (fd *FileSource) SourceType() string {
	return storageTypeLocalFile
}

func (fd *FileSource) Container() string {
	// needed for interface conformance
	return ""
}

func (fd *FileSource) GetSourceFilePath(path string) string {
	// needed for interface conformance
	return path
}

func (fd *FileSource) Reader(_ context.Context, path string) (io.Reader, error) {
	f, err := fd.FS.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return f, ErrSrcFileNotExist
		}
		return f, err
	}
	return f, nil
}

func (fd *FileSource) GetMetadata(_ context.Context, tuid string) (map[string]string, error) {
	f, err := fd.FS.Open(tuid + ".meta")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrSrcFileNotExist
		}
		return nil, err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var m map[string]string
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	fMain, err := fd.FS.Open(tuid)
	if err != nil {
		return nil, err
	}
	info, e := fMain.Stat()
	if e == nil {
		m["last_modified"] = info.ModTime().Format(time.RFC3339Nano)
		m["content_length"] = strconv.FormatInt(info.Size(), 10)
	}
	return m, nil
}

func (fd *FileSource) GetSignedObjectURL(_ context.Context, _ string, _ string) (string, error) {
	// needed for interface conformance
	return "", nil
}

func (fd *FileSource) Health(_ context.Context) (rsp models.ServiceHealthResp) {
	rsp.Service = "File Source"
	info, err := fs.Stat(fd.FS, ".")
	if err != nil {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = err.Error()
		return rsp
	}
	if !info.IsDir() {
		rsp.Status = models.STATUS_DOWN
		rsp.HealthIssue = fmt.Sprintf("%s is not a directory", info.Name())
		return rsp
	}
	rsp.Status = models.STATUS_UP
	rsp.HealthIssue = models.HEALTH_ISSUE_NONE
	return rsp
}
