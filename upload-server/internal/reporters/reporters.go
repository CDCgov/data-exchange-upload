package reporters

//
//import (
//	"context"
//	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
//	"io"
//	"log/slog"
//	"reflect"
//	"strings"
//)
//
//var logger *slog.Logger
//
//func init() {
//	type Empty struct{}
//	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
//	// add package name to app logger
//	logger = sloger.With("pkg", pkgParts[len(pkgParts)-1])
//}
//
//type Reporter interface {
//	io.Closer
//	Publish(context.Context, Identifiable) error
//}
