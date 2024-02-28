package server 

import (
	"net/http"
) // .import

func (s Server) version(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("%%VERSION%%")) // TODO: replace with git SHA in github actions at commit time

} // .version
