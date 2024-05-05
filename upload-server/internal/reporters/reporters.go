package reporters

import (
	"context"
)

type Identifiable interface {
	Identifier() string
}

type Reporter interface {
	Publish(context.Context, Identifiable) error
}
