package server 

import (
	"net/http"
)

func health(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("running ok"))

} // .health
