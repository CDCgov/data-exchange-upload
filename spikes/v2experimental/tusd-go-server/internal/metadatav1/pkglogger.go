package metadatav1

import (
	"log/slog"

	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
) // .import

// pkgLogger returns a custom package logger based on the app logger and added with this package name
func pkgLogger(appConfig appconfig.AppConfig) *slog.Logger {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := sloger.AppLogger(appConfig).With("pkg", pkgParts[len(pkgParts)-1])

	return logger
} // .pkgLogger