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
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"

	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

type ServerDex struct {
	cliFlags    cliflags.Flags
	appConfig   appconfig.AppConfig
	handlerTusd *tusd.Handler
	handlerDex  *handlerdex.HandlerDex
	logger      *slog.Logger
} // .ServerDex

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

	handlerDex, err := handlerdex.New(cliFlags, appConfig)
	if err != nil {
		logger.Error("error starting dex handler")
		return ServerDex{}, err
	} // .handlerDex

	return ServerDex{
		cliFlags:    cliFlags,
		appConfig:   appConfig,
		handlerTusd: handlerTusd,
		handlerDex:  handlerDex,
		logger: logger,
	}, nil // .return

} // New

func (sd *ServerDex) HttpServer() http.Server {

	// --------------------------------------------------------------
	// 		TUSD handler as-is
	// --------------------------------------------------------------

	// Right now, nothing has happened since we need to start the HTTP server on
	// our own. In the end, tusd will start listening on and accept request at
	// http://localhost:8080/files
	http.Handle("/files/", http.StripPrefix("/files/", sd.handlerTusd))

	// Start another goroutine for receiving events from the handler whenever
	// an upload is completed. The event will contains details about the upload
	// itself and the relevant HTTP request.
	go func() {
		for {
			event := <-sd.handlerTusd.CompleteUploads
			sd.logger.Info("upload finished", "event.Upload.ID", event.Upload.ID)
			// fmt.Printf("Upload %s finished\n", event.Upload.ID)
		}
	}() // .go func

	// --------------------------------------------------------------
	// 		DEX handler, handles all other requests except TUSD path
	// --------------------------------------------------------------
	http.Handle("/", sd.handlerDex)

	// --------------------------------------------------------------
	// 		Custom Server, if needed to customize
	// --------------------------------------------------------------
	return http.Server{

		Addr: sd.appConfig.ServerPort,
		// etc...

	} // .httpServer
} // .HttpServer
