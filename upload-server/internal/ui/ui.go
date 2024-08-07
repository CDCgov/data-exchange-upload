package ui

import (
	"context"
	"embed"
	"net/http"
)

// content holds our static web server content.
//
//go:embed assets/* index.html
var content embed.FS

var Handler = http.FileServer(http.FS(content))

var DefaultServer = NewServer(":8000")

func NewServer(addr string) *http.Server {
	s := &http.Server{
		Addr:    addr,
		Handler: Handler,
	}
	return s
}

func Start() error {
	return DefaultServer.ListenAndServe()
}

func Close(ctx context.Context) error {
	return DefaultServer.Shutdown(ctx)
}
