package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/coreos/go-oidc/v3/oidc"
)

func OAuthTokenVerificationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: check for specific scopes
		// TODO: opaque token

		ctx := context.Background()

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
		_, err = verifier.Verify(ctx, token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to verify token: %v", err), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
