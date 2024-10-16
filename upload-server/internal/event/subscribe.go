package event

import (
	"context"
	"io"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
)


type Subscribable[T Identifiable] interface {
	health.Checkable
	io.Closer
	GetBatch(ctx context.Context, max int) ([]T, error)
	HandleSuccess(ctx context.Context, event T) error
	HandleError(ctx context.Context, event T, handlerError error) error
}
