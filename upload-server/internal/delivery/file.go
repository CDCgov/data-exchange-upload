package delivery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

func NewFileDestination(_ context.Context, target string, appConfig *appconfig.AppConfig) (*FileDestination, error) {
	localConfig, err := appconfig.LocalStoreConfig(target, appConfig)
	if err != nil {
		return nil, err
	}

	return &FileDestination{
		LocalStorageConfig: *localConfig,
		Target:             target,
	}, nil
}

type FileDestination struct {
	appconfig.LocalStorageConfig
	Target string
}

func (fd *FileDestination) Upload(_ context.Context, id string, r io.Reader, m map[string]string) (string, error) {
	os.Mkdir(fd.ToPath, 0755)
	dest, err := os.Create(filepath.Join(fd.ToPath, id))
	if err != nil {
		return dest.Name(), err
	}
	defer dest.Close()
	if _, err := io.Copy(dest, r); err != nil {
		return dest.Name(), err
	}
	return dest.Name(), nil
}

type FileSource struct {
	FS fs.FS
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

	return m, nil
}
