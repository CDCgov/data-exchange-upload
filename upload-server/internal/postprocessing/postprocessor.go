package postprocessing

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
)

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

type PostProcessor struct {
	UploadBaseDir string
	UploadDir     string
}

func ProcessFileReadyEvent(ctx context.Context, e *event.FileReady) error {
	src, ok := delivery.GetSource("upload")
	if !ok {
		return fmt.Errorf("failed to get source for file delivery %+v", e)
	}
	d, ok := delivery.GetDestination(e.DestinationTarget)
	if !ok {
		return fmt.Errorf("failed to get destination for file delivery %+v", e)
	}
	return delivery.Deliver(ctx, e.UploadId, src, d)
}
