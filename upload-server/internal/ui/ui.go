package ui

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"html/template"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/eventials/go-tus"
)

// content holds our static web server content.
//
//go:embed assets/* index.html destination/* manifest.tmpl upload.tmpl
var content embed.FS

func FixNames(name string) string {
	removeChars := strings.ReplaceAll(name, "_", " ")
	newName := strings.Title(strings.ToLower(removeChars))
	return newName
}

var usefulFuncs = template.FuncMap{
	"FixNames": FixNames,
}

var manifestTemplate = template.Must(template.New("manifest.tmpl").Funcs(usefulFuncs).ParseFS(content, "manifest.tmpl"))
var uploadTemplate = template.Must(template.ParseFS(content, "upload.tmpl"))

var StaticHandler = http.FileServer(http.FS(content))

var DefaultServer = NewServer(":8000")

func NewServer(addr string) *http.Server {
	router := http.NewServeMux()
	router.HandleFunc("/manifest", func(rw http.ResponseWriter, r *http.Request) {
		// TODO check to see if they don't exist
		dataStream := r.FormValue("data_stream")
		dataStreamRoute := r.FormValue("data_stream_route")

		config, err := metadata.Cache.GetConfig(r.Context(), fmt.Sprintf("v2/%s-%s.json", dataStream, dataStreamRoute))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		err = manifestTemplate.Execute(rw, config)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	router.HandleFunc("/upload", func(rw http.ResponseWriter, r *http.Request) {
		// Tell the tus server we want to start an upload
		// turn form values into map[string]string
		err := r.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}
		manifest := map[string]string{"version": "2.0"}
		for k, v := range r.Form {
			manifest[k] = v[0]
		}

		// submit to upload server to get upload id
		upload := &tus.Upload{
			Metadata: manifest,
		}

		req, err := http.NewRequest("POST", "http://localhost:8080/files/", nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		req.Header.Set("Content-Length", "0")
		req.Header.Set("Upload-Metadata", upload.EncodedMetadata())
		req.Header.Set("Upload-Defer-Length", "1")
		req.Header.Set("Tus-Resumable", "1.0.0")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		loc := resp.Header.Get("Location")

		http.Redirect(rw, r, fmt.Sprintf("/status/%s", filepath.Base(loc)), 302)
	})
	router.HandleFunc("/status/{upload_id}", func(rw http.ResponseWriter, r *http.Request) {
		id := r.PathValue("upload_id")
		// TODO need to name this dynamic
		uploadUrl := "http://localhost:8080/files/" + id

		uploadTemplate.Execute(rw, uploadUrl)
	})
	router.Handle("/", StaticHandler)

	s := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	return s
}

func Start() error {
	return DefaultServer.ListenAndServe()
}

func Close(ctx context.Context) error {
	return DefaultServer.Shutdown(ctx)
}
