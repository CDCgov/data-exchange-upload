package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/coreos/go-oidc/v3/oidc"
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

func OAuthTokenVerificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		if len(authHeader) < len("Bearer ") {
			http.Error(w, "Authorization header format is invalid", http.StatusUnauthorized)
			return
		}

		token := authHeader[len("Bearer "):]

		var err error
		if strings.Count(token, ".") == 2 {
			// Token is JWT, validate using oidc verifier
			err = validateJWT(ctx, token)
		} else {
			// Token is opaque, validate using introspection
			err = validateOpaqueToken(ctx, token)
		}

		if err != nil {
			if httpErr, ok := err.(*HTTPError); ok {
				http.Error(w, httpErr.Msg, httpErr.Code)
			} else {
				http.Error(w, "unknown error occurred", http.StatusInternalServerError)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateJWT(ctx context.Context, token string) error {
	issuer := appconfig.LoadedConfig.OauthConfig.IssuerUrl

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get provider: %v", err))
	}

	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Failed to verify token: %v", err))
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return NewHTTPError(http.StatusUnauthorized, "Failed to parse token claims")
	}

	actualScopes := strings.Split(claims.Scopes, " ")
	requiredScopes := []string{}
	if appconfig.LoadedConfig.OauthConfig.RequiredScopes != "" {
		requiredScopes = strings.Split(appconfig.LoadedConfig.OauthConfig.RequiredScopes, " ")
	}

	if !hasRequiredScopes(actualScopes, requiredScopes) {
		return NewHTTPError(http.StatusForbidden, "One or more required scopes not found")
	}

	return nil
}

func validateOpaqueToken(ctx context.Context, token string) error {
	introspectionURL := appconfig.LoadedConfig.OauthConfig.IntrospectionUrl

	req, err := http.NewRequestWithContext(ctx, "POST", introspectionURL, strings.NewReader("token="+token))
	if err != nil {
		return NewHTTPError(http.StatusInternalServerError, "failed to create introspection request")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return NewHTTPError(http.StatusUnauthorized, "failed to validate opaque token")
	}
	defer resp.Body.Close()

	var introspectionResponse IntrospectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&introspectionResponse); err != nil {
		return NewHTTPError(http.StatusInternalServerError, "failed to parse introspection response")
	}

	if !introspectionResponse.Active {
		return NewHTTPError(http.StatusUnauthorized, "inactive token")
	}

	actualScopes := strings.Split(introspectionResponse.Scope, " ")
	requiredScopes := []string{}
	if appconfig.LoadedConfig.OauthConfig.RequiredScopes != "" {
		requiredScopes = strings.Split(appconfig.LoadedConfig.OauthConfig.RequiredScopes, " ")
	}

	if !hasRequiredScopes(actualScopes, requiredScopes) {
		return NewHTTPError(http.StatusForbidden, "one or more required scopes not found")
	}

	return nil
}

func hasRequiredScopes(actualScopes, requiredScopes []string) bool {
	if len(requiredScopes) == 0 {
		return true
	}

	for _, reqScope := range requiredScopes {
		if !slices.Contains(actualScopes, reqScope) {
			return false
		}
	}
	return true
}
