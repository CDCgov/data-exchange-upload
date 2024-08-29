package info

import "errors"

var (
	ErrNotFound = errors.New("expected file not found")
)

type InfoResponse struct {
	Manifest map[string]any `json:"manifest"`
	FileInfo map[string]any `json:"file_info"`
	FileStatus FileStatus `json:"file_status"`
}

type InfoFileData struct {
	MetaData map[string]any `json:"MetaData"`
}

type FileStatus struct {
	Destinations []FileDeliveryStatus `json:"destinations"`
}

type FileDeliveryStatus struct {
	Status string `json:"status"`
	Name string `json:"name"`
	Location string `json:"location"`
	DeliveredAt string `json:"delivered_at"`
}
