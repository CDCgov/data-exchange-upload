package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"slices"
	"sync/atomic"
	"time"
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
	if len(info.Deliveries) < len(c.ExpectedDeliveryTargets) {
		return &ErrAssertion{
			Expected: len(c.ExpectedDeliveryTargets),
			Actual:   len(info.Deliveries),
			msg:      "delivery count",
		}
	}

	if len(info.Deliveries) > len(c.ExpectedDeliveryTargets) {
		return errors.Join(&ErrAssertion{
			Expected: len(c.ExpectedDeliveryTargets),
			Actual:   len(info.Deliveries),
			msg:      "delivery count",
		}, &ErrFatalAssertion{
			msg: "delivered to more targets than expected",
		})
	}

	for _, delivery := range info.Deliveries {
		// Calculate time difference in milliseconds
		chunkReceivedAt, _ := time.Parse(time.RFC3339Nano, info.UploadStatus.LastChunkReceived)
		deliveredAt, _ := time.Parse(time.RFC3339Nano, delivery.DeliveredAt)
		timeDiff := deliveredAt.Sub(chunkReceivedAt).Milliseconds()

		// Log or add timeDiff to statistics
		addDeliveryTime(timeDiff) // new function to store delivery times

		if delivery.Status != "SUCCESS" || !slices.Contains(c.ExpectedDeliveryTargets, delivery.Name) {
			return fmt.Errorf("unexpected delivery status or target. upload_id=%s", uploadId)
		}
	}

	name := fmt.Sprintf("./output/manifests/%v.json", info.Manifest["upload_id"])
	err = os.WriteFile(name, b, 0644)
	if err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// Add delivery time to test result statistics
func addDeliveryTime(timeDiff int64) {
	atomic.AddInt64(&testResult.TotalDeliveryTime, timeDiff) // Accumulating time
	testResult.DeliveryTimes = append(testResult.DeliveryTimes, timeDiff)
	atomic.AddInt32(&testResult.DeliveryCount, 1)
}

func (ic *InfoChecker) OnSuccess() {
	atomic.AddInt32(&testResult.SuccessfulDeliveries, 1)
}

func (ic *InfoChecker) OnFail() error {
	return nil
}
