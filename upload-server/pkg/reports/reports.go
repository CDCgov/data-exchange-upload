package reports

import (
	"context"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
)

var Reporters event.Publishers[*Report]

func Register(r event.Publisher[*Report]) {
	Reporters = append(Reporters, r)
}

func Publish(ctx context.Context, r *Report) {
	Reporters.Publish(ctx, r)
}

func CloseAll() {
	Reporters.Close()
}
