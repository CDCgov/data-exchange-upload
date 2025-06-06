package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
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
	Client           *http.Client
	uploadComplete   time.Time
	deliveryComplete time.Time
}

func (ic *InfoChecker) DoCase(ctx context.Context, c TestCase, uploadId string) error {
	infoUrl, err := neturl.JoinPath(infoUrl, uploadId)
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
		if delivery.Status != "SUCCESS" {
			return errors.Join(&ErrAssertion{
				Expected: "SUCCESS",
				Actual:   delivery.Status,
			}, &ErrFatalAssertion{
				msg:      "unexpected delivery status",
				uploadId: uploadId,
			})
		}

		if !slices.Contains(c.ExpectedDeliveryTargets, delivery.Name) {
			return errors.Join(&ErrAssertion{
				Expected: c.ExpectedDeliveryTargets,
				Actual:   delivery.Name,
			}, &ErrFatalAssertion{
				msg:      "unexpected delivery target",
				uploadId: uploadId,
			})
		}
	}

	ic.uploadComplete, err = time.Parse(time.RFC3339Nano, info.UploadStatus.LastChunkReceived)
	if err != nil {
		return errors.Join(&ErrFatalAssertion{"error parsing upload complete timestamp", uploadId}, err)
	}
	ic.deliveryComplete, err = time.Parse(time.RFC3339Nano, info.Deliveries[0].DeliveredAt)
	if err != nil {
		return errors.Join(&ErrFatalAssertion{"error parsing delivery complete timestamp", uploadId}, err)
	}

	return nil
}

func (ic *InfoChecker) OnSuccess() {
	atomic.AddInt32(&testResult.SuccessfulDeliveries, 1)
	if !ic.uploadComplete.IsZero() && !ic.deliveryComplete.IsZero() {
		deliveryDuration := ic.deliveryComplete.Sub(ic.uploadComplete)
		testResult.DeliveryDurations = append(testResult.DeliveryDurations, deliveryDuration)
	}
}

func (ic *InfoChecker) OnFail() error {
	return nil
}
