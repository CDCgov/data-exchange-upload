package serverdex

import (
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	//
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlertusd"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/storecopier"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/hooks"
) // .import

// SeverDex, main Upload Api server, handles requests to both tusd handler and dex handler
type ServerDex struct {
	cliFlags    cliflags.Flags
	appConfig   appconfig.AppConfig
	handlerTusd *tusd.Handler
	handlerDex  *handlerdex.HandlerDex
	logger      *slog.Logger
} // .ServerDex

// New returns an custom server for DEX Upload Api ready to serve
func New(cliFlags cliflags.Flags, appConfig appconfig.AppConfig) (ServerDex, error) {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := sloger.AppLogger(appConfig).With("pkg", pkgParts[len(pkgParts)-1])

	handlerTusd, err := handlertusd.New(cliFlags, appConfig)
	if err != nil {
		logger.Error("error starting tusd handler")
		return ServerDex{}, err
	} // .handlerTusd

	handlerDex := handlerdex.New(cliFlags, appConfig)

	return ServerDex{
		cliFlags:    cliFlags,
		appConfig:   appConfig,
		handlerTusd: handlerTusd,
		handlerDex:  handlerDex,
		logger:      logger,
	}, nil // .return

} // New

// HttpServer, adds the routes for the tusd and dex handlers and can customize the server with port address
func (sd *ServerDex) HttpServer() http.Server {

	// --------------------------------------------------------------
	// 	TUSD handler
	// --------------------------------------------------------------
	// Route for TUSD to start listening on and accept http request
	http.Handle(sd.appConfig.TusdHandlerBasePath, http.StripPrefix(sd.appConfig.TusdHandlerBasePath, sd.handlerTusd))

	// Start another goroutine for receiving events from the handler whenever
	// an upload is completed. The event will contains details about the upload
	// itself and the relevant HTTP request.
	go func() {
		for {

			event := <-sd.handlerTusd.CompleteUploads
			sd.logger.Info("upload finished", "event.Upload.ID", event.Upload.ID)

			fileName := event.Upload.MetaData["filename"]
			// copy A -> B
			stA := storecopier.StoreLocal{
				FileLocalFolder: sd.appConfig.LocalFolderUploads,
				FileName:        event.Upload.ID,
			} // .a
			stB := storecopier.StoreLocal{
				FileLocalFolder: sd.appConfig.LocalFolderUploadsB,
				FileName:        fileName,
			} // .b
			err := storecopier.CopySrcToDst(stA, stB)
			if err != nil {
				sd.logger.Error("error copy A -> B", "error", err)
			} else {
				sd.logger.Info("copied file A -> B ok")
			} // .else
			// copy B -> BC
			stC := storecopier.StoreLocal{
				FileLocalFolder: sd.appConfig.LocalFolderUploadsC,
				FileName:        fileName,
			} // .a
			err = storecopier.CopySrcToDst(stB, stC)
			if err != nil {
				sd.logger.Error("error copy B -> C", "error", err)
			} else {
				sd.logger.Info("copied file B -> C ok")
			} // .else

		} // .for
	}() // .go func

	// --------------------------------------------------------------
	// 	Prometheus metrics handler for /metrics
	// --------------------------------------------------------------
	sd.setupMetrics(sd.handlerTusd)
	hooks.SetupHookMetrics()
	http.Handle("/metrics", promhttp.Handler())

	// --------------------------------------------------------------
	// 	DEX handler for all other http requests except above
	// --------------------------------------------------------------
	http.Handle("/", sd.handlerDex)

	// --------------------------------------------------------------
	// 		Custom Server, if needed to customize
	// --------------------------------------------------------------
	return http.Server{

		Addr: ":" + sd.appConfig.ServerPort,
		// etc...

	} // .httpServer
} // .HttpServer
