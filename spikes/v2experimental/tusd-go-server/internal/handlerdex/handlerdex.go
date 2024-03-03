package handlerdex

import (
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/cliflags"
	"net/http"
) // .import

type HandlerDex struct {
	flags  cliflags.Flags
	config appconfig.AppConfig
}

func New(flags cliflags.Flags, config appconfig.AppConfig) (*HandlerDex, error) {

	return &HandlerDex{flags: flags, config: config}, nil
} // .New

func (hd HandlerDex) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {

	case "/":
		hd.root(w, r)

	case "/health":
		hd.health(w, r)

	case "/version":
		hd.version(w, r)

	// all other non-specified routes
	default:
		http.NotFound(w, r)

	} // .switch

} // .ServeHTTP
