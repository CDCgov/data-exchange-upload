package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"path"
	"slices"
	"sync/atomic"
)

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

type InfoChecker struct {
	Client *http.Client
}

func (ic *InfoChecker) DoCase(ctx context.Context, c TestCase, uploadId string) error {
	serverUrl, _ := path.Split(url)
	infoUrl, err := neturl.JoinPath(serverUrl, "info", uploadId)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", infoUrl, nil)
	if err != nil {
		return err
	}

	resp, err := ic.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to check upload info: %d, %s", resp.StatusCode, infoUrl)
	}

	var info InfoResponse
	b, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &info)
	if err != nil {
		return err
	}

	// Check info fields
	if info.UploadStatus.Status != "Complete" {
		return fmt.Errorf("upload unsuccessful for upload ID %s: %v", uploadId, info.UploadStatus)
	}

	// Check delivery targets
	if len(info.Deliveries) != len(c.ExpectedDeliveryTargets) {
		return fmt.Errorf("expected %d deliveries but got %d", len(c.ExpectedDeliveryTargets), len(info.Deliveries))
	}

	for _, delivery := range info.Deliveries {
		if delivery.Status != "SUCCESS" {
			return errors.Join(&ErrAssertion{
				Expected: "SUCCESS",
				Actual:   delivery.Status,
			}, &ErrFatalAssertion{"unexpected delivery status"})
		}

		if !slices.Contains(c.ExpectedDeliveryTargets, delivery.Name) {
			return errors.Join(&ErrAssertion{
				Expected: c.ExpectedDeliveryTargets,
				Actual:   delivery.Name,
			}, &ErrFatalAssertion{"unexpected delivery target"})
		}
	}

	return nil
}

func (ic *InfoChecker) OnSuccess() {
	atomic.AddInt32(&testResult.SuccessfulDeliveries, 1)
}

func (ic *InfoChecker) OnFail() error {
	return nil
}
