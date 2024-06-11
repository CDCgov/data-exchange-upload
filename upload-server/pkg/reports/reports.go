package reports

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/reporters"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"log/slog"
	"reflect"
	"strings"
)

var logger *slog.Logger
var DefaultReporter reporters.Reporter

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

func Publish(ctx context.Context, r reporters.Identifiable) {
	if err := DefaultReporter.Publish(ctx, r); err != nil {
		logger.Error("Failed to report", "report", r, "reporter", DefaultReporter, "err", err)
	}
}
