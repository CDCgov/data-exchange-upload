package cli

import (
	"context"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
)

type UploadStatusInspector interface {
	InspectFileDeliveryStatus(ctx context.Context, id string) ([]info.FileDeliveryStatus, error)
	InspectFileUploadStatus(ctx context.Context, id string) (info.FileUploadStatus, error)
}
