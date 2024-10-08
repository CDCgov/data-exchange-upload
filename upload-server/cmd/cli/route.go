package cli

import (
	"encoding/json"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/metadata"
	"io"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/delivery"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/event"
)

type Router struct{}
type RequestBody struct {
	DataStreamId    string `json:"data_stream_id"`
	DataStreamRoute string `json:"data_stream_route"`
	Target          string `json:"target"`
	Source          string `json:"source"`
}

func (router *Router) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	id := r.PathValue("UploadID")
	b, err := io.ReadAll(r.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	var body RequestBody
	err = json.Unmarshal(b, &body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	sourceName := body.Source
	if sourceName == "" {
		sourceName = "upload"
	}
	src, ok := delivery.GetSource(sourceName)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Could not find source"))
		return
	}

	m, err := src.GetMetadata(r.Context(), id)
	if err != nil {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte(err.Error()))
		return
	}

	dataStreamId, dataStreamRoute := metadata.GetDataStreamID(m), metadata.GetDataStreamRoute(m)
	if _, ok := delivery.GetDestinationTarget(dataStreamId, dataStreamRoute, body.Target); !ok {
		http.Error(rw, "Invalid target", http.StatusBadRequest)
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
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
		return
	}

	rw.Write([]byte(id))
}
