package ui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	v2 "github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/v2"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata/validation"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/ui/components"
	"github.com/cdcgov/data-exchange-upload/upload-server/pkg/info"

	"html/template"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/metadata"
	"github.com/dustin/go-humanize"
	"github.com/eventials/go-tus"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// content holds our static web server content.
//
//go:embed assets/* components/* index.html manifest.html upload.html
var content embed.FS

func FixNames(name string) string {
	removeChars := strings.ReplaceAll(name, "_", " ")

	caser := cases.Title(language.English)
	newName := caser.String(removeChars)

	return newName
}

func KebabCase(value string) string {
	kebabValue := strings.ReplaceAll(value, " ", "-")
	return strings.ToLower(kebabValue)
}

func FormatDateTime(dateTimeString string) string {
	if dateTimeString == "" {
		return ""
	}

	date, err := time.Parse(time.RFC3339, dateTimeString)

	if err != nil {
		return ""
	}

	return date.UTC().Format(time.RFC850)
}

func FormatBytes(bytes float64) string {
	intBytes := uint64(bytes)
	strBytes := humanize.Bytes(intBytes)

	return strings.ToUpper(strBytes)
}

var usefulFuncs = template.FuncMap{
	"FixNames":       FixNames,
	"AllLowerCase":   strings.ToLower,
	"AllUpperCase":   strings.ToUpper,
	"FormatDateTime": FormatDateTime,
	"KebabCase":      KebabCase,
	"FormatBytes":    FormatBytes,
}

func generateTemplate(templatePath string, useFuncs bool) *template.Template {
	var templatePaths = []string{templatePath, "components/navbar.html", "components/newuploadbtn.html"}
	if useFuncs {
		return template.Must(template.New(templatePath).Funcs(usefulFuncs).ParseFS(content, templatePaths...))
	}
	return template.Must(template.ParseFS(content, templatePaths...))
}

var indexTemplate = generateTemplate("index.html", false)
var manifestTemplate = generateTemplate("manifest.html", true)
var uploadTemplate = generateTemplate("upload.html", true)

type ManifestTemplateData struct {
	DataStream      string
	DataStreamRoute string
	MetadataFields  []validation.FieldConfig
	Navbar          components.Navbar
	CsrfToken       string
}

type UploadTemplateData struct {
	UploadUrl    string
	UploadStatus string
	Info         info.InfoResponse
	Navbar       components.Navbar
	NewUploadBtn components.NewUploadBtn
}

var StaticHandler = http.FileServer(http.FS(content))

func NewServer(addr string, csrfToken string, uploadUrl string, infoUrl string) *http.Server {
	router := GetRouter(uploadUrl, infoUrl)
	secureRouter := csrf.Protect(
		[]byte(csrfToken),
		csrf.Secure(false), // TODO: make dynamic when supporting TLS
	)(router)

	s := &http.Server{
		Addr:    addr,
		Handler: secureRouter,
	}
	return s
}

func GetRouter(uploadUrl string, infoUrl string) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/manifest", func(rw http.ResponseWriter, r *http.Request) {
		dataStream := r.FormValue("data_stream_id")
		dataStreamRoute := r.FormValue("data_stream_route")

		configId := v2.ConfigIdentification{
			DataStreamID: dataStream,
			DataStreamRoute: dataStreamRoute,
		}

		config, err := metadata.Cache.GetConfig(r.Context(), configId.Path())
		if err != nil {
			http.Error(rw, err.Error(), http.StatusNotFound)
			return
		}

		err = manifestTemplate.Execute(rw, &ManifestTemplateData{
			DataStream:      dataStream,
			DataStreamRoute: dataStreamRoute,
			MetadataFields:  filterMetadataFields(config),
			Navbar:          components.NewNavbar(false),
			CsrfToken:       csrf.Token(r),
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
		uuid := resp.Header.Get("Location")
		uuid = filepath.Base(uuid)

		http.Redirect(rw, r, fmt.Sprintf("/status/%s", uuid), http.StatusFound)
	}).Methods("POST")
	router.HandleFunc("/status/{upload_id}", func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["upload_id"]

		// Check for upload
		u, err := url.Parse(infoUrl)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		u = u.JoinPath(id)
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		// Redirect to landing page if info not found
		if resp.StatusCode == http.StatusNotFound {
			http.Redirect(rw, r, "/", http.StatusFound)
			return
		}

		// Get the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		if resp.StatusCode != http.StatusOK {
			http.Error(rw, string(body), resp.StatusCode)
			return
		}

		var fileInfo info.InfoResponse
		err = json.Unmarshal(body, &fileInfo)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		uploadUrl, err := url.JoinPath(uploadUrl, id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = uploadTemplate.Execute(rw, &UploadTemplateData{
			UploadUrl:    uploadUrl,
			Info:         fileInfo,
			UploadStatus: fileInfo.UploadStatus.Status,
			Navbar:       components.NewNavbar(true),
			NewUploadBtn: components.NewUploadBtn{},
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	router.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		err := indexTemplate.Execute(rw, nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.PathPrefix("/assets/").Handler(StaticHandler)

	return router
}

var DefaultServer *http.Server

func Start(uiPort string, csrfToken string, uploadURL string, infoURL string) error {
	DefaultServer = NewServer(uiPort, csrfToken, uploadURL, infoURL)

	return DefaultServer.ListenAndServe()
}

func Close(ctx context.Context) error {
	if DefaultServer != nil {
		return DefaultServer.Shutdown(ctx)
	}
	return nil
}

func filterMetadataFields(config *validation.ManifestConfig) []validation.FieldConfig {
	var fields []validation.FieldConfig

	for _, f := range config.Metadata.Fields {
		if f.FieldName != "data_stream_id" && f.FieldName != "data_stream_route" {
			fields = append(fields, f)
		}
	}

	return fields
}
