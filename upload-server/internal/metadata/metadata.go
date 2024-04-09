package metadata

import (
	"log/slog"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
)

var logger *slog.Logger

func init() {
	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
}

func VerifySenderManifest(event handler.HookEvent) (hooks.HookResponse, error) {
	manifest := event.Upload.MetaData
	logger.Info("checking the sender manifest:", "manifest", manifest)
	resp := hooks.HookResponse{}
	return resp, nil
}
