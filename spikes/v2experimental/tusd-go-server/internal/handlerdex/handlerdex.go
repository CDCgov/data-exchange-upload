package handlerdex

import (
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/pkg/sloger"
) // .import

type HandlerDex struct {
	cliFlags  cliflags.Flags
	appConfig appconfig.AppConfig
	logger    *slog.Logger

	// azure service dependencies
	TusAzBlobClient    *azblob.Client
	RouterAzBlobClient *azblob.Client
	EdavAzBlobClient   *azblob.Client
} // .HandlerDex

// New returns a DEX sever handler that can handle http requests
func New(flags cliflags.Flags, appConfig appconfig.AppConfig) *HandlerDex {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := sloger.AppLogger(appConfig).With("pkg", pkgParts[len(pkgParts)-1])

	logger.Info("started dex handler")

	return &HandlerDex{
		cliFlags:  flags,
		appConfig: appConfig,
		logger:    logger,
	} // .&HandlerDex
} // .New

// ServeHTTP handler method for handling http requests
// it routes request for response to different paths to respective functions
func (hd HandlerDex) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {

	case "/":
		hd.root(w, r)

	case "/health":
		hd.health(w, r)

	case "/version":
		hd.version(w, r)

	case "/metadata":
		hd.metadata(w, r)

	case "/metadata/v1":
		hd.metadataV1(w, r)

	// all other non-specified routes
	default:
		http.NotFound(w, r)

	} // .switch

} // .ServeHTTP
