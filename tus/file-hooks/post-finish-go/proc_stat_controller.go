package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ProcStatController struct {
	URL        string
	DelayS     time.Duration
	Client     *http.Client
	Logger     *log.Logger
	RetryCount int
	MaxRetries int
}

func NewProcStatController(url string, delayS time.Duration) *ProcStatController {
	return &ProcStatController{
		URL:        url,
		DelayS:     delayS,
		Client:     &http.Client{},
		Logger:     log.New(os.Stdout, "ProcStatController: ", log.LstdFlags),
		RetryCount: 0,
		MaxRetries: 6, // or fetch from environment if dynamic control is required
	}
}

func (p *ProcStatController) GetSpanByUploadID(uploadID, stageName string) (traceID, spanID string, err error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/trace/span?uploadId=%s&stageName=%s", p.URL, uploadID, stageName), nil)
	if err != nil {
		return "", "", err
	}

	resp, err := p.sendRequestWithRetry(req)
	if err != nil {
		return "", "", err
	}

	var result struct {
		TraceID string `json:"trace_id"`
		SpanID  string `json:"span_id"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", err
	}
	if result.TraceID == "" || result.SpanID == "" {
		return "", "", fmt.Errorf("invalid PS API response: %+v", result)
	}

	return result.TraceID, result.SpanID, nil
}

func (p *ProcStatController) StopSpanForTrace(traceID, parentSpanID string) error {
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/trace/stopSpan/%s/%s", p.URL, traceID, parentSpanID), nil)
	if err != nil {
		return err
	}

	_, err = p.sendRequestWithRetry(req)
	return err
}

func (p *ProcStatController) sendRequestWithRetry(req *http.Request) (*http.Response, error) {
    for p.RetryCount = 0; p.RetryCount < p.MaxRetries; p.RetryCount++ {
        resp, err := p.Client.Do(req)
        if err != nil {
            p.Logger.Printf("Error sending request to PS API after attempt %d.  Reason: %v", p.RetryCount, err)
            time.Sleep(p.DelayS)
            continue
        }
        defer resp.Body.Close()

        if resp.StatusCode == http.StatusOK {
            return resp, nil
        }

        if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable {
            return nil, fmt.Errorf("non-retryable status code %d received", resp.StatusCode)
        }

        // Extracting and handling the Retry-After header
        retryAfterHeader := resp.Header.Get("Retry-After")
        if retryAfterHeader != "" {
            retryAfter, err := strconv.Atoi(retryAfterHeader)
            if err != nil {
                p.Logger.Printf("Failed to parse Retry-After header: %v", err)
                time.Sleep(p.DelayS)
            } else {
                retryAfterDuration := time.Duration(retryAfter) * time.Second
                time.Sleep(retryAfterDuration)
            }
        } else {
            time.Sleep(p.DelayS)
        }
    }
    return nil, fmt.Errorf("unable to send successful request to PS API after %d attempts", p.MaxRetries)
}


