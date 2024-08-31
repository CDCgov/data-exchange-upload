package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/coreos/go-oidc/v3/oidc"
)

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

		if strings.Count(token, ".") == 2 {
			// Token is JWT, validate using oidc verifier
			if !validateJWT(ctx, w, token) {
				return
			}
		} else {
			// Token is opaque, validate using introspection
			if !validateOpaqueToken(ctx, w, token) {
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func validateJWT(ctx context.Context, w http.ResponseWriter, token string) bool {
	issuer := appconfig.LoadedConfig.OauthIssuerUrl

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		http.Error(w, "Failed to get provider", http.StatusUnauthorized)
		return false
	}

	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify token: %v", err), http.StatusUnauthorized)
		return false
	}

	var claims struct {
		Scopes string `json:"scope"`
	}

	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse token claims: %v", err), http.StatusUnauthorized)
		return false
	}

	actualScopes := strings.Split(claims.Scopes, " ")

	requiredScopes := []string{}
	if appconfig.LoadedConfig.OauthRequiredScopes != "" {
		requiredScopes = strings.Split(appconfig.LoadedConfig.OauthRequiredScopes, " ")
	}

	if !hasRequiredScopes(actualScopes, requiredScopes) {
		http.Error(w, "One or more required scopes not found.", http.StatusForbidden)
		return false
	}

	return true
}

func validateOpaqueToken(ctx context.Context, w http.ResponseWriter, token string) bool {
	introspectionURL := appconfig.LoadedConfig.OauthIntrospectionUrl

	req, err := http.NewRequestWithContext(ctx, "POST", introspectionURL, strings.NewReader("token="+token))
	if err != nil {
		http.Error(w, "Failed to create introspection request", http.StatusInternalServerError)
		return false
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Optionally add Basic Auth if required by the IdP
	// req.SetBasicAuth(appconfig.LoadedConfig.ClientID, appconfig.LoadedConfig.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to validate opaque token", http.StatusUnauthorized)
		return false
	}
	defer resp.Body.Close()

	var introspectionResponse struct {
		Active bool   `json:"active"`
		Scope  string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&introspectionResponse); err != nil {
		http.Error(w, "Failed to parse introspection response", http.StatusInternalServerError)
		return false
	}

	if !introspectionResponse.Active {
		http.Error(w, "Inactive token", http.StatusUnauthorized)
		return false
	}

	actualScopes := strings.Split(introspectionResponse.Scope, " ")

	requiredScopes := []string{}
	if appconfig.LoadedConfig.OauthRequiredScopes != "" {
		requiredScopes = strings.Split(appconfig.LoadedConfig.OauthRequiredScopes, " ")
	}

	if !hasRequiredScopes(actualScopes, requiredScopes) {
		http.Error(w, "One or more required scopes not found.", http.StatusForbidden)
		return false
	}

	return true
}

func hasRequiredScopes(actualScopes, requiredScopes []string) bool {
	if len(requiredScopes) == 0 {
		return true
	}

	scopeMap := make(map[string]bool)
	for _, scope := range actualScopes {
		scopeMap[scope] = true
	}
	for _, reqScope := range requiredScopes {
		if !scopeMap[reqScope] {
			return false
		}
	}
	return true
}
