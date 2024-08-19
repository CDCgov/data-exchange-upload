package ui

import (
	"context"
	"embed"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"html/template"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/eventials/go-tus"
)

// content holds our static web server content.
//
//go:embed assets/* index.html manifest.tmpl upload.tmpl
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

type ManifestTemplateData struct {
	DataStream      string
	DataStreamRoute string
	MetadataFields  []validation.FieldConfig
}

var StaticHandler = http.FileServer(http.FS(content))

func NewServer(addr string, uploadUrl string) *http.Server {
	s := &http.Server{
		Addr:    addr,
		Handler: GetRouter(uploadUrl),
	}
	return s
}

func GetRouter(uploadUrl string) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("/manifest", func(rw http.ResponseWriter, r *http.Request) {
		dataStream := r.FormValue("data_stream")
		dataStreamRoute := r.FormValue("data_stream_route")

		config, err := metadata.Cache.GetConfig(r.Context(), fmt.Sprintf("v2/%s-%s.json", dataStream, dataStreamRoute))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		err = manifestTemplate.Execute(rw, &ManifestTemplateData{
			DataStream:      dataStream,
			DataStreamRoute: dataStreamRoute,
			MetadataFields:  metadata.GetMetadataFields(config),
		})
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

		req, err := http.NewRequest("POST", uploadUrl, nil)
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
		if resp.StatusCode != http.StatusCreated {
			// Failed to init upload.  Forward response from tus.
			var respMsg []byte
			_, err := resp.Body.Read(respMsg)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Error(rw, string(respMsg), resp.StatusCode)
			return
		}
		loc := resp.Header.Get("Location")
		http.Redirect(rw, r, fmt.Sprintf("/status/%s", filepath.Base(loc)), http.StatusFound)
	})
	router.HandleFunc("/status/{upload_id}", func(rw http.ResponseWriter, r *http.Request) {
		id := r.PathValue("upload_id")
		uploadUrl, err := url.JoinPath(uploadUrl, id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = uploadTemplate.Execute(rw, uploadUrl)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	router.Handle("/", StaticHandler)

	return router
}

var DefaultServer *http.Server

func Start(uiPort string, uploadURL string) error {
	DefaultServer = NewServer(uiPort, uploadURL)
	return DefaultServer.ListenAndServe()
}

func Close(ctx context.Context) error {
	return DefaultServer.Shutdown(ctx)
}
