package server 

import (
	"net/http"
)

func (s Server) health(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("running ok"))

} // .health
