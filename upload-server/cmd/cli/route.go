package cli

import (
	"encoding/json"
	"errors"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/postprocessing"
	"io"
	"net/http"
)

type Router struct{}
type RequestBody struct {
	Target string `json:"target"`
}

func (router *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		rw.WriteHeader(400)
		return
	}
	id := r.PathValue("UploadID")
	b, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(400)
		return
	}
	var body RequestBody
	err = json.Unmarshal(b, &body)
	if err != nil {
		rw.WriteHeader(400)
		return
	}

	err = postprocessing.Deliver(r.Context(), id, nil, body.Target)
	if err != nil {
		if errors.Is(err, postprocessing.ErrBadTarget) {
			rw.WriteHeader(400)
			rw.Write([]byte(err.Error() + " " + body.Target))
			return
		}
		if errors.Is(err, postprocessing.ErrSrcFileNotExist) {
			rw.WriteHeader(404)
			rw.Write([]byte(err.Error()))
			return
		}

		// Unhandled error occurred
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write([]byte(id))
}
