package handlertusd

import (
	"reflect"
	"strings"

	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
	"github.com/tus/tusd/v2/pkg/filestore"
	tusd "github.com/tus/tusd/v2/pkg/handler"
)

func New(cliFlags cliflags.Flags, appConfig appconfig.AppConfig) (*tusd.Handler, error) {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := sloger.AppLogger(appConfig).With("pkg", pkgParts[len(pkgParts)-1])

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
		// added

		// TODO: 	the tusd logger type is "golang.org/x/exp/slog" vs. app logger "log/slog" ?
		// TODO: pass the custom app logger
		// Logger: logger,

		PreUploadCreateCallback: checkManifestV1,
	}) // .handler
	if err != nil {
		logger.Error("error start tusd handler", "error", err)
		return nil, err
	} // .if

	logger.Info("started tusd handler")
	return handler, nil
} // .New
