package serverdex

import (
	"fmt"
	"net/http"

	//
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlertusd"

	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

type ServerDex struct {
	cliFlags    cliflags.Flags
	appConfig   appconfig.AppConfig
	handlerTusd *tusd.Handler
	handlerDex  *handlerdex.HandlerDex
} // .ServerDex

func New(cliFlags cliflags.Flags, appConfig appconfig.AppConfig) (ServerDex, error) {

	handlerTusd, err := handlertusd.New(cliFlags, appConfig)
	if err != nil {
		return ServerDex{}, err
	} // .handlerTusd

	handlerDex, err := handlerdex.New(cliFlags, appConfig)
	if err != nil {
		return ServerDex{}, err
	} // .handlerDex

	return ServerDex{
		cliFlags:    cliFlags,
		appConfig:   appConfig,
		handlerTusd: handlerTusd,
		handlerDex:  handlerDex,
	}, nil // .return

} // New

func (s *ServerDex) HttpServer() http.Server {

	// --------------------------------------------------------------
	// 		TUSD handler as-is
	// --------------------------------------------------------------

	// Right now, nothing has happened since we need to start the HTTP server on
	// our own. In the end, tusd will start listening on and accept request at
	// http://localhost:8080/files
	http.Handle("/files/", http.StripPrefix("/files/", s.handlerTusd))

	// Start another goroutine for receiving events from the handler whenever
	// an upload is completed. The event will contains details about the upload
	// itself and the relevant HTTP request.
	go func() {
		for {
			event := <-s.handlerTusd.CompleteUploads
			fmt.Printf("Upload %s finished\n", event.Upload.ID)
		}
	}() // .go func

	// --------------------------------------------------------------
	// 		DEX handler, handles all other requests except TUSD path
	// --------------------------------------------------------------
	http.Handle("/", s.handlerDex)

	// --------------------------------------------------------------
	// 		Custom Server, if needed to customize
	// --------------------------------------------------------------
	return http.Server{

		Addr: s.appConfig.ServerPort,
		// etc...

	} // .httpServer
} // .HttpServer
