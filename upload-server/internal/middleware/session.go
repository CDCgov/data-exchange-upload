package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
)

var store sessions.Store

func InitStore(key string) {
	store = sessions.NewCookieStore([]byte(key))
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
	if s.IsNew {
		// set security flags for newly created session
		s.Options = &sessions.Options{
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
	}

	return &UserSession{s}, nil
}

func (s *UserSession) Data() UserSessionData {
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
