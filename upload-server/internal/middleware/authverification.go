package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/health"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/session"
	"github.com/gorilla/sessions"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/oauth"
)

var ErrNoAuthHeader = errors.New("authorization header missing")
var ErrAuthHeaderInvalidFormat = errors.New("authorization header format is invalid")
var ErrTokenNotFound = errors.New("authorization token not found")

const UserSessionCookieName = "phdo_session"
const LoginRedirectCookieName = "login_redirect"

type Claims struct {
	Scopes string `json:"scope"`
}

type IntrospectionResponse struct {
	Active bool   `json:"active"`
	Scope  string `json:"scope"`
}

type HTTPError struct {
	Code int
	Msg  string
}

func (e *HTTPError) Error() string {
	return e.Msg
}

func NewHTTPError(code int, msg string) *HTTPError {
	return &HTTPError{Code: code, Msg: msg}
}

func NewAuthMiddleware(ctx context.Context, config appconfig.OauthConfig) (*AuthMiddleware, error) {
	var validator oauth.Validator = oauth.PassthroughValidator{}
	if config.AuthEnabled {
		if config.IssuerUrl == "" {
			return nil, errors.New("no issuer url provided")
		}
		var err error
		validator, err = oauth.NewOAuthValidator(ctx, config.IssuerUrl, config.RequiredScopes)
		if err != nil {
			slog.Error("error initializing oauth validator", "error", err)
			return nil, err
		}
		health.Register(validator)
	}

	return &AuthMiddleware{
		authEnabled: config.AuthEnabled,
		validator:   validator,
	}, nil
}

type AuthMiddleware struct {
	authEnabled bool
	validator   oauth.Validator
}

func (a AuthMiddleware) VerifyOAuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.authEnabled {
			next.ServeHTTP(w, r)
			return
		}
		// allow preflight checks from browser clients
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		// read auth token from either headers or cookies
		token, err := getAuthToken(r.Header)
		if err != nil {
			if errors.Is(err, ErrNoAuthHeader) {
				// fallback to cookies
				token = getAuthTokenFromCookies(*r)
			} else {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
		}
		if token == "" {
			http.Error(w, ErrTokenNotFound.Error(), http.StatusUnauthorized)
			return
		}
		if strings.Count(token, ".") == 2 {
			// Token is JWT, validate using oidc verifier
			_, err = a.validator.ValidateJWT(r.Context(), token)
			if err != nil {
				if errors.Is(err, oauth.ErrTokenVerificationFailed) || errors.Is(err, oauth.ErrTokenClaimsFailed) {
					err = errors.Join(err, NewHTTPError(http.StatusUnauthorized, err.Error()))
				} else if errors.Is(err, oauth.ErrTokenScopesMismatch) {
					err = errors.Join(err, NewHTTPError(http.StatusForbidden, err.Error()))
				} else {
					err = errors.Join(err, NewHTTPError(http.StatusInternalServerError, err.Error()))
				}
			}
		} else {
			// Token is opaque, validate using introspection
			err = validateOpaqueToken()
		}
		if err != nil {
			var httpErr *HTTPError
			if errors.As(err, &httpErr) {
				http.Error(w, httpErr.Msg, httpErr.Code)
			}
			slog.Warn("request failed token validation", "path", r.URL.Path, "error", httpErr.Msg)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a AuthMiddleware) VerifyUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.authEnabled {
			next.ServeHTTP(w, r)
			return
		}

		sess, _ := session.Store().Get(r, UserSessionCookieName)
		token := sess.Values["token"]
		if token == nil {
			loginRedirect(sess, w, r)
			return
		}

		_, err := a.validator.ValidateJWT(r.Context(), token.(string))
		if err != nil {
			loginRedirect(sess, w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a AuthMiddleware) Validator() oauth.Validator {
	return a.validator
}

func loginRedirect(sess *sessions.Session, w http.ResponseWriter, r *http.Request) {
	if sess != nil {
		v := r.URL.Path
		if r.URL.RawQuery != "" {
			v += "?" + r.URL.RawQuery
		}
		sess.Values["redirect"] = v
	}
	sess.Save(r, w)

	//http.SetCookie(w, &http.Cookie{
	//	Name:     LoginRedirectCookieName,
	//	Value:    v,
	//	Path:     "/oauth_callback",
	//	Secure:   true,
	//	HttpOnly: true,
	//	SameSite: http.SameSiteStrictMode,
	//})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func validateOpaqueToken() error {
	// TODO: work out opaque token validation logic
	return nil
}

func getAuthToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeader
	}

	if len(authHeader) < len("Bearer ") {
		return "", ErrAuthHeaderInvalidFormat
	}

	return authHeader[len("Bearer "):], nil
}

func getAuthTokenFromCookies(r http.Request) string {
	c, err := r.Cookie(UserSessionCookieName)
	if err != nil && errors.Is(err, http.ErrNoCookie) {
		return ""
	}
	return c.Value
}

//func redirectSanitized(path string, code int, w http.ResponseWriter, r *http.Request) {
//	redirectURL := path
//	if redirectURL == "" {
//		// default to home page
//		redirectURL = "/"
//	}
//	// build target URL
//	targetURL := r.URL.Path
//	if r.URL.RawQuery != "" {
//		// append incoming query params
//		targetURL += "?" + r.URL.RawQuery
//	}
//
//	if targetURL != "" {
//		redirectURL += "?redirect=" + url.QueryEscape(targetURL)
//	}
//
//	if isValidRedirectURL(redirectURL) {
//		http.Redirect(w, r, redirectURL, code)
//		return
//	}
//	http.Error(w, "invalid redirect url", http.StatusBadRequest)
//}

//func sanitizeRedirectPath(path string, r http.Request) string {
//	sanitized := path
//
//	//append redirect query
//	var redirectQuery string
//	redirectPath := r.URL.Path
//	if r.URL.RawQuery != "" {
//		redirectPath += "?" + r.URL.RawQuery
//	}
//	sanitizedRedirect := sanitizeRedirectUrl(redirectPath)
//	if sanitizedRedirect != "/" {
//		redirectQuery = "?redirect=" + sanitizedRedirect
//	}
//
//	sanitized += redirectQuery
//
//	return sanitized
//}

//func isValidRedirectURL(redirectURL string) bool {
//	if redirectURL == "" {
//		return false
//	}
//
//	parsed, err := url.Parse(redirectURL)
//	if err != nil {
//		return false
//	}
//
//	if parsed.IsAbs() || !strings.HasPrefix(parsed.Path, "/") {
//		return false
//	}
//
//	return true
//}

//func sanitizeRedirectUrl(redirectURL string) string {
//	sanitized := "/"
//
//	if redirectURL == "" {
//		return sanitized
//	}
//
//	parsed, err := url.Parse(redirectURL)
//	if err != nil {
//		return sanitized
//	}
//
//	if parsed.IsAbs() || !strings.HasPrefix(parsed.Path, "/") {
//		return sanitized
//	}
//
//	//for _, p := range protectedUIRoutes {
//	//	if strings.HasPrefix(parsed.Path, p) {
//	//		sanitized = parsed.Path
//	//		if parsed.RawQuery != "" {
//	//			sanitized += "?" + url.QueryEscape(parsed.RawQuery)
//	//		}
//	//		return sanitized
//	//	}
//	//}
//
//	return sanitized
//}
