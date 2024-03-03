package handlertusd

import (
	"log/slog"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/config"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/flags"
	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

func New(flags flags.Flags, config config.Config) (*tusd.Handler, error) {

	// Create a new FileStore instance which is responsible for
	// storing the uploaded file on disk in the specified directory.
	// This path _must_ exist before tusd will store uploads in it.
	// If you want to save them on a different medium, for example
	// a remote FTP server, you can implement your own storage backend
	// by implementing the tusd.DataStore interface.
	store := filestore.FileStore{
		Path: "./uploads",
	} // .store

	// A storage backend for tusd may consist of multiple different parts which
	// handle upload creation, locking, termination and so on. The composer is a
	// place where all those separated pieces are joined together. In this example
	// we only use the file store but you may plug in multiple.
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	// Create a new HTTP handler for the tusd server by providing a configuration.
	// The StoreComposer property must be set to allow the handler to function.
	handler, err := tusd.NewHandler(tusd.Config{
		BasePath:              "/files/",
		StoreComposer:         composer,
		NotifyCompleteUploads: true,
		//
		PreUploadCreateCallback: checkManifestV1,
	}) // .handler

	if err != nil {
		slog.Error("tushandler: unable to create new tusd handler", "error", err)
		return nil, err
	} // .if

	// no error
	return handler, nil
} // .New
