package handlerdex

import (
	"log/slog"
	"net/http"
	"reflect"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/processingstatus"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/sloger"
) // .import

type HandlerDex struct {
	appConfig appconfig.AppConfig
	logger    *slog.Logger

	// azure service dependencies
	TusAzBlobClient    *azblob.Client
	RouterAzBlobClient *azblob.Client
	EdavAzBlobClient   *azblob.Client

	// processing status
	PsSender *processingstatus.PsSender
} // .HandlerDex

// New returns a DEX sever handler that can handle http requests
func New(appConfig appconfig.AppConfig, psSender *processingstatus.PsSender) *HandlerDex {

	type Empty struct{}
	pkgParts := strings.Split(reflect.TypeOf(Empty{}).PkgPath(), "/")
	// add package name to app logger
	logger := sloger.With("pkg", pkgParts[len(pkgParts)-1])

	logger.Info("started dex handler")

	return &HandlerDex{
		appConfig: appConfig,
		logger:    logger,
		PsSender:  psSender,
	} // .&HandlerDex
} // .New

// ServeHTTP handler method for handling http requests
// it routes request for response to different paths to respective functions
// TODO: refactor to use a servmux
func (hd HandlerDex) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {

	case "/":
		appconfig.Handler().ServeHTTP(w, r)

	case "/health":
		health.Handler().ServeHTTP(w, r)

	case "/version":
		hd.version(w, r)

	case "/metadata":
		hd.metadata(w, r)

	// all other non-specified routes
	default:
		http.NotFound(w, r)

	} // .switch

} // .ServeHTTP
