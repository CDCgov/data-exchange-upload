package reports

import (
	"context"
	"log/slog"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
)

var logger *slog.Logger
var DefaultReporter event.Publisher[*Report]
var Reporters []event.Publisher[*Report]

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

func Register(r event.Publisher[*Report]) {
	Reporters = append(Reporters, r)
}

func Publish(ctx context.Context, r *Report) {
	for _, reporter := range Reporters {
		err := reporter.Publish(ctx, r)
		if err != nil {
			logger.Error("Failed to report", "report", r, "reporter", reporter, "err", err)
			if r.RetryCount() < event.MaxRetries {
				r.IncrementRetryCount()
				Publish(ctx, r)
			}
		}
	}
}

func CloseAll() {
	for _, r := range Reporters {
		r.Close()
	}
	Reporters = []event.Publisher[*Report]{}
}
