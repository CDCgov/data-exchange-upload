package main

type InfoResponse struct {
	Manifest     map[string]any       `json:"manifest"`
	FileInfo     map[string]any       `json:"file_info"`
	UploadStatus FileUploadStatus     `json:"upload_status"`
	Deliveries   []FileDeliveryStatus `json:"deliveries"`
}

type FileUploadStatus struct {
	Status            string `json:"status"`
	LastChunkReceived string `json:"chunk_received_at"`
}

type FileDeliveryStatus struct {
	Status      string  `json:"status"`
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	DeliveredAt string  `json:"delivered_at"`
	Issues      []Issue `json:"issues"`
}

type Issue struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}
