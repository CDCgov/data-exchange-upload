package info

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("expected file not found")
)

const (
	UploadInitiated string = "Initiated"
	UploadInProgress string = "In Progress"
	UploadComplete string = "Complete"
)

type InfoResponse struct {
	Manifest   map[string]any       `json:"manifest"`
	FileInfo   map[string]any       `json:"file_info"`
	UploadStatus FileUploadStatus `json:"upload_status"`
	Deliveries []FileDeliveryStatus `json:"deliveries"`
}

type InfoFileData struct {
	MetaData map[string]any `json:"MetaData"`
}

type FileUploadStatus struct {
	Status string `json:"status"`
	LastChunkReceived time.Time `json:"chunk_received_at"`
}

type FileDeliveryStatus struct {
	Status      string   `json:"status"`
	Name        string   `json:"name"`
	Location    string   `json:"location"`
	DeliveredAt string   `json:"delivered_at"`
	Issues      []string `json:"issues"`
}
