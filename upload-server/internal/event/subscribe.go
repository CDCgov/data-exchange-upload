package event

import (
	"context"
	"io"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
)

type Subscribable[T Identifiable] interface {
	health.Checkable
	io.Closer
	Listen(context.Context, func(context.Context, T) error) error
}
