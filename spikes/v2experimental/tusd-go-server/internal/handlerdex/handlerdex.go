package handlerdex

import (
	"net/http"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/config"
	"github.com/cdcgov/data-exchange-upload/tusd-go-server/internal/flags"
) // .import 

type HandlerDex struct {

}

func New(flags flags.Flags, config config.Config) (*HandlerDex, error) {


	return &HandlerDex{}, nil 
} // .New

func (hd HandlerDex) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    switch r.URL.Path {
    case "/":
        hd.root(w, r)

    case "/health":
        hd.health(w, r)

	case "/version":
        hd.version(w, r)

	// all non-specific routes
    default:
        http.NotFound(w, r)
    } // .default 

} // .ServeHTTP
