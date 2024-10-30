package event

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
)

const TypeSeparator = "_"

type FilePublisher[T Identifiable] struct {
	Dir string
}

func (mp *FilePublisher[T]) Publish(_ context.Context, event T) error {
	err := os.MkdirAll(mp.Dir, 0750)
	if err != nil && !os.IsExist(err) {
		return err
	}

	filename := filepath.Join(mp.Dir, event.Identifier()+TypeSeparator+event.Type())
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// write event to file.
	encoder := json.NewEncoder(f)
	err = encoder.Encode(event)
	if err != nil {
		return err
	}

	return nil
}
