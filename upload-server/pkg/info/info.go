package info

import "errors"

var (
	ErrNotFound = errors.New("expected file not found")
)

const (
	UploadNotStarted string = "Not Started"
	UploadInProgress string = "In Progress"
	UploadComplete string = "Complete"
	UploadFailed string = "Failed"

)

type InfoResponse struct {
	Manifest map[string]any `json:"manifest"`
	FileInfo   map[string]any `json:"file_info"`
	UploadStatus string `json:"upload_status"`
	Deliveries []FileDeliveryStatus `json:"deliveries"`
}

type InfoFileData struct {
	MetaData map[string]any `json:"MetaData"`
}

type FileDeliveryStatus struct {
	Status string `json:"status"`
	Name string `json:"name"`
	Location string `json:"location"`
	DeliveredAt string `json:"delivered_at"`
	Issues []string `json:"issues"`
}
