package event

import (
	"context"
)

type Subscribable[T Identifiable] interface {
	Listen(context.Context, func(context.Context, T) error) error
}
