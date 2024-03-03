package server

import (
	"fmt"
	"net/http"

	//
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/config"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/flags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlerdex"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/handlertusd"

	tusd "github.com/tus/tusd/v2/pkg/handler"
) // .import

type Server struct {
	flags       flags.Flags
	config      config.Config
	handlerTusd *tusd.Handler
	handlerDex *handlerdex.HandlerDex
} // .Server

func New(flags flags.Flags, config config.Config, handlerTusd *tusd.Handler) (Server, error) {

	handlerTusd, err := handlertusd.New(flags, config)
	if err != nil {
		return Server{}, err
	} // .handlerTusd

	handlerDex, err := handlerdex.New(flags, config)
	if err != nil {
		return Server{}, err 
	} // .handlerDex

	return Server{
		flags:       flags,
		config:      config,
		handlerTusd: handlerTusd,
		handlerDex: handlerDex,
	}, nil // .return

} // New

func (s Server) Serve() error {

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
	customServer := http.Server{
		Addr: s.config.ServerPort,
	} // .customServer


	return customServer.ListenAndServe()
} // .Serve
