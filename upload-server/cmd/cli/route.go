package cli

import (
	"encoding/json"
	"errors"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
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

	d, ok := postprocessing.GetTarget(body.Target)
	if !ok {
		rw.WriteHeader(400)
		rw.Write([]byte("invalid target " + body.Target))
		return
	}
	srcUrl, err := d.GetSrcUrl(r.Context(), id)
	if err != nil {
		if errors.Is(err, postprocessing.ErrSrcFileNotExist) {
			rw.WriteHeader(404)
			rw.Write([]byte(err.Error()))
			return
		}
	}
	m, err := d.GetMetadata(r.Context(), id)
	if err != nil {
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	e := &event.FileReady{
		Event: event.Event{
			Type: event.FileReadyEventType,
		},
		UploadId:          id,
		DestinationTarget: body.Target,
		Metadata:          m,
		SrcUrl:            srcUrl,
	}
	err = event.FileReadyPublisher.Publish(r.Context(), e)
	if err != nil {
		// Unhandled error occurred
		rw.WriteHeader(500)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write([]byte(id))
}
