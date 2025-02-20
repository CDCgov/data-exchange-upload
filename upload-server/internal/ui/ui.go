package ui

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/middleware"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/oauth"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
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
//go:embed assets/* components/* index.html manifest.html upload.html login.html
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

	date, err := time.Parse(time.RFC3339Nano, dateTimeString)

	if err != nil {
		date, err = time.Parse(time.RFC3339, dateTimeString)
		if err != nil {
			return ""
		}
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
	var templatePaths = []string{templatePath, "components/navbar.html", "components/linkbtn.html"}
	if useFuncs {
		return template.Must(template.New(templatePath).Funcs(usefulFuncs).ParseFS(content, templatePaths...))
	}
	return template.Must(template.ParseFS(content, templatePaths...))
}

var indexTemplate = generateTemplate("index.html", false)
var loginTemplate = generateTemplate("login.html", false)
var manifestTemplate = generateTemplate("manifest.html", true)
var uploadTemplate = generateTemplate("upload.html", true)

type LoginTemplateData struct {
	AuthFailed bool
	Redirect   string
	CsrfToken  string
}

type IndexTemplateData struct {
	Navbar components.Navbar
}

type ManifestTemplateData struct {
	DataStream      string
	DataStreamRoute string
	MetadataFields  []validation.FieldConfig
	Navbar          components.Navbar
	CsrfToken       string
	AuthEnabled     bool
	AuthFailed      bool
}

type UploadTemplateData struct {
	UploadEndpoint string
	UploadUrl      string
	UploadStatus   string
	Info           info.InfoResponse
	Navbar         components.Navbar
	NewUploadBtn   components.LinkBtn
}

var StaticHandler = http.FileServer(http.FS(content))

func NewServer(ctx context.Context, port string, csrfToken string, externalUploadUrl string, externalInfoUrl string, internalUploadUrl string, authConfig appconfig.OauthConfig) (*http.Server, error) {
	var validator oauth.Validator = oauth.PassthroughValidator{}
	if authConfig.AuthEnabled {
		var err error
		validator, err = oauth.NewOAuthValidator(ctx, authConfig.IssuerUrl, authConfig.RequiredScopes)
		if err != nil {
			//logger.Error("error initializing oauth validator", "error", err)
			return nil, err
		}
	}
	//oauthValidator, err := oauth.NewOAuthValidator(ctx, authConfig.IssuerUrl, authConfig.RequiredScopes)
	//if err != nil {
	//	return nil, err
	//}
	authMiddleware := middleware.NewAuthMiddleware(validator, authConfig.AuthEnabled)

	router := GetRouter(externalUploadUrl, externalInfoUrl, internalUploadUrl, authMiddleware)
	secureRouter := csrf.Protect(
		[]byte(csrfToken),
		csrf.Secure(false), // TODO: make dynamic when supporting TLS
	)(router)

	addr := fmt.Sprintf(":%s", port)

	s := &http.Server{
		Addr:    addr,
		Handler: secureRouter,
	}
	return s, nil
}

func GetRouter(externalUploadUrl string, internalInfoUrl string, internalUploadUrl string, authMiddleware middleware.AuthMiddleware) *mux.Router {
	router := mux.NewRouter()
	protectedRouter := router.PathPrefix("/").Subrouter()
	protectedRouter.Use(authMiddleware.VerifyUserSession)

	router.HandleFunc("/login", func(rw http.ResponseWriter, r *http.Request) {
		redirect := r.URL.Query().Get("redirect")
		authFailed, err := strconv.ParseBool(r.FormValue("auth_failed"))
		if err != nil {
			authFailed = false
		}

		err = loginTemplate.Execute(rw, &LoginTemplateData{
			AuthFailed: authFailed,
			Redirect:   redirect,
			CsrfToken:  csrf.Token(r),
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	router.HandleFunc("/logout", func(rw http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(middleware.UserSessionCookieName)
		if err != nil {
			http.Redirect(rw, r, "/login", http.StatusFound)
			return
		}
		c.Expires = time.Unix(0, 0)
		c.MaxAge = -1
		http.SetCookie(rw, c)

		http.Redirect(rw, r, "/login", http.StatusFound)
	})
	router.HandleFunc("/oauth_callback", func(rw http.ResponseWriter, r *http.Request) {
		redirect := r.URL.Query().Get("redirect")
		if redirect == "" {
			redirect = "/"
		}
		if !strings.HasPrefix(redirect, "/") {
			redirect = "/" + redirect
		}
		token := r.FormValue("token")

		if token == "" {
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}

		claims, err := authMiddleware.Validator().ValidateJWT(r.Context(), token)
		if err != nil {
			http.Redirect(rw, r, "/login", http.StatusSeeOther)
			return
		}

		http.SetCookie(rw, &http.Cookie{
			Name:     middleware.UserSessionCookieName,
			Value:    token,
			Path:     "/",
			Expires:  time.Unix(claims.Expiry, 0),
			MaxAge:   int(claims.Expiry),
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(rw, r, redirect, http.StatusFound)
	}).Methods("POST")
	protectedRouter.HandleFunc("/manifest", func(rw http.ResponseWriter, r *http.Request) {
		dataStream := r.FormValue("data_stream_id")
		dataStreamRoute := r.FormValue("data_stream_route")

		configId := metadata.ConfigIdentification{
			DataStreamID:    dataStream,
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
			Navbar:          components.NewNavbar(false, isLoggedIn(*r)),
			CsrfToken:       csrf.Token(r),
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	protectedRouter.HandleFunc("/upload", func(rw http.ResponseWriter, r *http.Request) {
		// Tell the tus server we want to start an upload
		// turn form values into map[string]string
		err := r.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			return
		}

		r.Form.Del("gorilla.csrf.Token")
		manifest := map[string]string{"version": "2.0"}
		for k, v := range r.Form {
			manifest[k] = v[0]
		}

		// submit to upload server to get upload id
		upload := &tus.Upload{
			Metadata: manifest,
		}

		req, err := http.NewRequest("POST", internalUploadUrl, nil)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Length", "0")
		req.Header.Set("Upload-Metadata", upload.EncodedMetadata())
		req.Header.Set("Upload-Defer-Length", "1")
		req.Header.Set("Tus-Resumable", "1.0.0")
		req.Header.Set("Authorization", r.Header.Get("Authorization"))

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO handle status unauthorized

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
	protectedRouter.HandleFunc("/status/{upload_id}", func(rw http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["upload_id"]

		// Check for upload
		u, err := url.Parse(internalInfoUrl)
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

		req.Header.Set("Authorization", r.Header.Get("Authorization"))

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

		uploadDestinationUrl, err := url.JoinPath(externalUploadUrl, id)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		err = uploadTemplate.Execute(rw, &UploadTemplateData{
			UploadEndpoint: externalUploadUrl,
			UploadUrl:      uploadDestinationUrl,
			Info:           fileInfo,
			UploadStatus:   fileInfo.UploadStatus.Status,
			Navbar:         components.NewNavbar(true, isLoggedIn(*r)),
			NewUploadBtn:   components.LinkBtn{Href: "/", Text: "Upload New File"},
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	protectedRouter.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		err := indexTemplate.Execute(rw, &IndexTemplateData{
			Navbar: components.NewNavbar(false, isLoggedIn(*r)),
		})
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	router.PathPrefix("/assets/").Handler(StaticHandler)

	return router
}

var DefaultServer *http.Server

func Start(ctx context.Context, uiPort string, csrfToken string, externalUploadURL string, internalInfoURL string, internalUploadUrl string, authConfig appconfig.OauthConfig) error {
	DefaultServer, err := NewServer(ctx, uiPort, csrfToken, externalUploadURL, internalInfoURL, internalUploadUrl, authConfig)
	if err != nil {
		return err
	}

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

func isLoggedIn(r http.Request) bool {
	_, err := r.Cookie(middleware.UserSessionCookieName)
	if err != nil && errors.Is(err, http.ErrNoCookie) {
		return false
	}
	return true
}
