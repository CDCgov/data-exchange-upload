package reporters

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters"
)

type FileReporter struct {
	Dir string
}

func (fr *FileReporter) Publish(_ context.Context, r reporters.Identifiable) error {
	if fr.Dir != "" {
		err := os.Mkdir(fr.Dir, 0750)
		if err != nil && !os.IsExist(err) {
			return err
		}
	}
	filename := fr.Dir + "/" + r.Identifier()
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()
	encoder := json.NewEncoder(f)
	err = encoder.Encode(r)
	if err != nil {
		return err
	}

	return nil
}

func (fr *FileReporter) Close() error {
	return nil
}
