package postprocessing

import (
	"context"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"log/slog"
	"reflect"
	"strings"
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

type Event struct {
	ID       string
	Manifest map[string]string
	Target   string
}

func Worker(ctx context.Context, c chan Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case e := <-c:
			if err := Deliver(ctx, e.ID, e.Manifest, e.Target); err != nil {
				// TODO Retry
				logger.Error("error delivering file to target", "event", e, "error", err.Error())
			}
		}
	}
}
