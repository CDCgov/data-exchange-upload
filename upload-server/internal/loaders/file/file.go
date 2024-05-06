package file

import (
	"context"
	"errors"
	"io"
	"io/fs"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
)

type FileConfigLoader struct {
	FileSystem fs.FS
}

func (l *FileConfigLoader) LoadConfig(ctx context.Context, path string) ([]byte, error) {

	file, err := l.FileSystem.Open(path)
	if err != nil {
		return nil, errors.Join(err, validation.ErrNotFound)
	}
	return io.ReadAll(file)
}
