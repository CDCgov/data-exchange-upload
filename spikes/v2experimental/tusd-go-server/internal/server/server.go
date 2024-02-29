package server

import (
	"fmt"
	"net/http"
	//
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/config"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/flags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/tusdhandler"

	tusd "github.com/tus/tusd/v2/pkg/handler"

) // .import



type Server struct {
	flags flags.Flags
	config config.Config
	tusdHandler *tusd.Handler
}

func New(flags flags.Flags, config config.Config, tusdHandler *tusd.Handler) (Server, error) {

	tusdHandler, err := tusdhandler.New(flags, config)
	if err != nil {
		return Server{}, err 
	}

	return Server {
		flags: flags,
		config: config,
		tusdHandler: tusdHandler,
	}, nil // .return

} // New

func (s Server) Serve() error {

	// Right now, nothing has happened since we need to start the HTTP server on
	// our own. In the end, tusd will start listening on and accept request at
	// http://localhost:8080/files
	http.Handle("/files/", http.StripPrefix("/files/", s.tusdHandler))

	// Start another goroutine for receiving events from the handler whenever
	// an upload is completed. The event will contains details about the upload
	// itself and the relevant HTTP request.
	go func() {
		for {
			event := <-s.tusdHandler.CompleteUploads
			fmt.Printf("Upload %s finished\n", event.Upload.ID)
		}
	}() // .go func


	http.HandleFunc("/health", s.health)
	http.HandleFunc("/version", s.version)
	http.HandleFunc("/metadata/v1", s.configMetaV1)

	return http.ListenAndServe(s.config.ServerPort, nil)

} // .Serve