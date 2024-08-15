package serverdex

import (
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metrics"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
) // .import

// ServerDex, main Upload Api server, handles requests to both tusd handler and dex handler
type ServerDex struct {
	AppConfig appconfig.AppConfig

	logger *slog.Logger
} // .ServerDex

// New returns an custom server for DEX Upload Api ready to serve
func New(appConfig appconfig.AppConfig) (ServerDex, error) {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := sloger.With("pkg", pkgParts[len(pkgParts)-1])
	sloger.SetDefaultLogger(logger)

	return ServerDex{
		AppConfig: appConfig,
		logger:    logger,
	}, nil // .return

} // New

// HttpServer, adds the routes for the tusd and dex handlers and can customize the server with port address
func (sd *ServerDex) HttpServer() http.Server {

	// --------------------------------------------------------------
	// 		Custom Server, if needed to customize
	// --------------------------------------------------------------
	return http.Server{

		Addr: ":" + sd.AppConfig.ServerPort,

		Handler: metrics.TrackHTTP(http.DefaultServeMux),
		// etc...

	} // .httpServer
} // .HttpServer
