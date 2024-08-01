package cli

import "net/http"

type Router struct{}

func (router *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		rw.WriteHeader(400)
		return
	}
	id := r.PathValue("UploadID")
	rw.Write([]byte(id))
}
