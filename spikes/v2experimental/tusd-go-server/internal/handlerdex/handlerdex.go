package handlerdex

import (
	"net/http"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/config"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/flags"
) // .import 

type HandlerDex struct {
    flags flags.Flags
    config config.Config
}

func New(flags flags.Flags, config config.Config) (*HandlerDex, error) {

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
