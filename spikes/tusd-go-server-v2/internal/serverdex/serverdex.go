package serverdex

import (
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus"
) // .import

var metricsOpenConnections = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "dex_server_connections_open",
	Help: "Current number of server open connections.",
}) // .metricsOpenConnections

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

	prometheus.MustRegister(metricsOpenConnections)
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
