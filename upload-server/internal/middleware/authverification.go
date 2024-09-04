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
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateJWT(ctx context.Context, token string) error {
	issuer := appconfig.LoadedConfig.OauthConfig.IssuerUrl

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return fmt.Errorf("failed to get provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{SkipClientIDCheck: true})

	idToken, err := verifier.Verify(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}

	var claims Claims

	if err := idToken.Claims(&claims); err != nil {
		return fmt.Errorf("failed to parse token claims: %w", err)
	}

	actualScopes := strings.Split(claims.Scopes, " ")

	requiredScopes := []string{}
	if appconfig.LoadedConfig.OauthConfig.RequiredScopes != "" {
		requiredScopes = strings.Split(appconfig.LoadedConfig.OauthConfig.RequiredScopes, " ")
	}

	if !hasRequiredScopes(actualScopes, requiredScopes) {
		return fmt.Errorf("one or more required scopes not found")
	}

	return nil
}

func validateOpaqueToken(ctx context.Context, token string) error {
	introspectionURL := appconfig.LoadedConfig.OauthConfig.IntrospectionUrl

	req, err := http.NewRequestWithContext(ctx, "POST", introspectionURL, strings.NewReader("token="+token))
	if err != nil {
		return fmt.Errorf("failed to create introspection request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Optionally add Basic Auth if required by the IdP
	// req.SetBasicAuth(appconfig.LoadedConfig.ClientID, appconfig.LoadedConfig.ClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to validate opaque token")
	}
	defer resp.Body.Close()

	var introspectionResponse IntrospectionResponse

	if err := json.NewDecoder(resp.Body).Decode(&introspectionResponse); err != nil {
		return fmt.Errorf("failed to parse introspection response: %w", err)
	}

	if !introspectionResponse.Active {
		return fmt.Errorf("inactive token")
	}

	actualScopes := strings.Split(introspectionResponse.Scope, " ")

	requiredScopes := []string{}
	if appconfig.LoadedConfig.OauthConfig.RequiredScopes != "" {
		requiredScopes = strings.Split(appconfig.LoadedConfig.OauthConfig.RequiredScopes, " ")
	}

	if !hasRequiredScopes(actualScopes, requiredScopes) {
		return fmt.Errorf("one or more required scopes not found")
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
