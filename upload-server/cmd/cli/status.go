package cli

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"
)

type UploadStatusInspector interface {
	InspectFileStatus(ctx context.Context, id string) (*info.FileStatus, error)
}