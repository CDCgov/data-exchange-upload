package serverdex

import (
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/metadatav1"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/processingstatus"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/tus/tusd/v2/pkg/hooks"
) // .import

// ServerDex, main Upload Api server, handles requests to both tusd handler and dex handler
type ServerDex struct {
	AppConfig appconfig.AppConfig
	MetaV1    *metadatav1.MetadataV1

	HandlerDex *handlerdex.HandlerDex
	logger     *slog.Logger
} // .ServerDex

// New returns an custom server for DEX Upload Api ready to serve
func New(appConfig appconfig.AppConfig, metaV1 *metadatav1.MetadataV1, psSender *processingstatus.PsSender) (ServerDex, error) {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := sloger.With("pkg", pkgParts[len(pkgParts)-1])
	sloger.SetDefaultLogger(logger)

	handlerDex := handlerdex.New(appConfig, psSender)

	return ServerDex{
		AppConfig:  appConfig,
		MetaV1:     metaV1,
		HandlerDex: handlerDex,
		logger:     logger,
	}, nil // .return

} // New

// HttpServer, adds the routes for the tusd and dex handlers and can customize the server with port address
func (sd *ServerDex) HttpServer() http.Server {

	// --------------------------------------------------------------
	// 	Prometheus metrics handler for /metrics
	// --------------------------------------------------------------
	sd.setupMetrics()
	hooks.SetupHookMetrics()
	http.Handle("/metrics", promhttp.Handler())

	// --------------------------------------------------------------
	// 	DEX handler for all other http requests except above
	// --------------------------------------------------------------
	http.Handle("/", sd.HandlerDex)

	// --------------------------------------------------------------
	// 		Custom Server, if needed to customize
	// --------------------------------------------------------------
	return http.Server{

		Addr: ":" + sd.AppConfig.ServerPort,

		ConnState: func(_ net.Conn, cs http.ConnState) {
			switch cs {
			case http.StateNew:
				metricsOpenConnections.Inc()
			case http.StateClosed, http.StateHijacked:
				metricsOpenConnections.Dec()
			} // .switch
		},
		// etc...

	} // .httpServer
} // .HttpServer
