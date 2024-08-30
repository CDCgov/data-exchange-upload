package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/coreos/go-oidc/v3/oidc"
)

func OAuthTokenVerificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: opaque token

		ctx := r.Context()

		issuer := appconfig.LoadedConfig.OauthIssuerUrl

		// Create a provider using the IdP's issuer URL
		provider, err := oidc.NewProvider(ctx, issuer)
		if err != nil {
			http.Error(w, "Failed to get provider", http.StatusUnauthorized)
			return
		}

		// Set up the OIDC verifier
		verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		if len(authHeader) < len("Bearer ") {
			http.Error(w, "Authorization header format is invalid", http.StatusUnauthorized)
			return
		}

		// Extract the token from the header
		token := authHeader[len("Bearer "):]

		// Verify the token
		idToken, err := verifier.Verify(ctx, token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to verify token: %v", err), http.StatusUnauthorized)
			return
		}

		var claims struct {
			Scopes string `json:"scope"`
		}

		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse token claims: %v", err), http.StatusUnauthorized)
			return
		}
		actualScopes := strings.Split(claims.Scopes, " ")

		requiredScopes := []string{}
		if appconfig.LoadedConfig.OauthRequiredScopes != "" {
			requiredScopes = strings.Split(appconfig.LoadedConfig.OauthRequiredScopes, " ")
		}

		if !hasRequiredScopes(actualScopes, requiredScopes) {
			http.Error(w, "One or more required scopes not found.", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hasRequiredScopes(actualScopes, requiredScopes []string) bool {
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
