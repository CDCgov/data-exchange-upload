package info

import "errors"

var (
	ErrNotFound = errors.New("expected file not found")
)

type InfoResponse struct {
	Manifest map[string]any `json:"manifest"`
	FileInfo map[string]any `json:"file_info"`
}

type InfoFileData struct {
	MetaData map[string]any `json:"MetaData"`
}
