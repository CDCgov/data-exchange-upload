package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/gorilla/sessions"
)

var store sessions.Store

func InitStore(config appconfig.OauthConfig) error {
	if config.AuthEnabled && config.SessionKey == "" {
		return errors.New("no session key provided")
	}
	store = sessions.NewCookieStore([]byte(config.SessionKey))
	store.(*sessions.CookieStore).Options = &sessions.Options{
		Path:     "/",
		Secure:   config.SessionSecure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Domain:   "cdc.gov",
	}
	return nil
}

type UserSessionData struct {
	Token    string
	Redirect string
}

type UserSession struct {
	session *sessions.Session
}

func GetUserSession(r *http.Request) (*UserSession, error) {
	s, err := store.Get(r, UserSessionCookieName)
	if err != nil {
		return &UserSession{s}, err
	}

	return &UserSession{s}, nil
}

func (s *UserSession) Data() UserSessionData {
	slog.Info("session", "token", s.session.Values["token"])
	token, ok := s.session.Values["token"].(string)
	if !ok {
		token = ""
	}
	redirect, ok := s.session.Values["redirect"].(string)
	if !ok {
		redirect = ""
	}

	return UserSessionData{token, redirect}
}

func (s *UserSession) SetToken(r *http.Request, w http.ResponseWriter, token string, expiry int) error {
	s.session.Values["token"] = token
	s.session.Options.MaxAge = expiry
	return s.session.Save(r, w)
}
func (s *UserSession) SetRedirect(r *http.Request, w http.ResponseWriter, redirect string) error {
	s.session.Values["redirect"] = redirect
	return s.session.Save(r, w)
}
func (s *UserSession) Delete(r *http.Request, w http.ResponseWriter) error {
	s.session.Options.MaxAge = -1
	return s.session.Save(r, w)
}
