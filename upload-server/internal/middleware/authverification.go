package middleware

import (
	"errors"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/oauth"
	"net/http"
	"strings"
)

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

func NewAuthMiddleware(validator oauth.OAuthValidator, enabled bool) AuthMiddleware {
	return AuthMiddleware{
		authEnabled: enabled,
		validator:   validator,
	}
}

type AuthMiddleware struct {
	authEnabled bool
	validator   oauth.OAuthValidator
}

func (a AuthMiddleware) VerifyOAuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.authEnabled {
			next.ServeHTTP(w, r)
			return
		}

		token, err := getAuthToken(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if strings.Count(token, ".") == 2 {
			// Token is JWT, validate using oidc verifier
			err = a.validator.ValidateJWT(r.Context(), token)
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
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a AuthMiddleware) ProtectUIRouteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !a.authEnabled {
			next.ServeHTTP(w, r)
			return
		}

		_, err := getAuthToken(r.Header)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	})
}

func validateOpaqueToken() error {
	// TODO: work out opaque token validation logic
	return nil
}

func getAuthToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header missing")
	}

	if len(authHeader) < len("Bearer ") {
		return "", errors.New("authorization header format is invalid")
	}

	return authHeader[len("Bearer "):], nil
}
