package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
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

	/*
		srcUrl, err := d.GetSrcUrl(r.Context(), id)
		if err != nil {
			if errors.Is(err, delivery.ErrSrcFileNotExist) {
				rw.WriteHeader(404)
				rw.Write([]byte(err.Error()))
				return
			}
		}
	*/
	fromPathStr := appconfig.LoadedConfig.LocalFolderUploadsTus + "/" + appconfig.LoadedConfig.TusUploadPrefix
	fromPath := os.DirFS(fromPathStr)
	src := &delivery.FileSource{
		FS: fromPath,
	}

	m, err := src.GetMetadata(r.Context(), id)
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
		SrcUrl:            id,
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
