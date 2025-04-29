package logutil

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

func TestWithUploadId(t *testing.T) {
	var logOutput bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logOutput, nil))
	slog.SetDefault(logger)
	expected := "test-upload-id"

	// Create a mock event and response
	event := &handler.HookEvent{
		Upload: handler.FileInfo{
			ID: expected,
		},
		Context: context.Background(),
	}
	resp := hooks.HookResponse{}

	// Call the function
	_, err := WithUploadIdLogger(event, resp)

	// Check for errors
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check if the upload ID is set in the context
	if event.Context.Value(sloger.LoggerKey) == nil {
		t.Fatalf("expected upload ID to be set in context")
	}

	if !strings.Contains(logOutput.String(), expected) {
		t.Fatalf("expected log output to contain %s, got %s", expected, logOutput.String())
	}
}
