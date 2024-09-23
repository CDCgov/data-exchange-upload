package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

// TestOAuthTokenVerificationMiddleware_AuthIsDisabled()
//
//	tests the OAuthTokenVerificationMiddleware function when
//	the AuthEnabled flag is false (Disabled).
func TestOAuthTokenVerificationMiddleware_AuthIsDisabled(t *testing.T) {
	// Save & restore the orig config
	originalConfig := appconfig.LoadedConfig
	defer func() { appconfig.LoadedConfig = originalConfig }()

	// Init the OauthConfig with AuthEnabled set to false
	appconfig.LoadedConfig = &appconfig.AppConfig{
		OauthConfig: &appconfig.OauthConfig{
			AuthEnabled:    false,
			IssuerUrl:      "https://issuer.example.com",
			RequiredScopes: "",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	hasBeenCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasBeenCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := OAuthTokenVerificationMiddleware(handler)
	middleware.ServeHTTP(rec, req)

	if !hasBeenCalled {
		t.Fatal("expected handler to not be called when auth is disabled")
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected a not OK status, got %v", rec.Code)
	}
}

// TestOAuthTokenVerificationMiddleware_AuthIsEnabled()
//
//	tests the OAuthTokenVerificationMiddleware function
//	when the AuthEnabled flag is true (Enabled).
func TestOAuthTokenVerificationMiddleware_AuthIsEnabled(t *testing.T) {
	// Save & restore the orig config
	originalConfig := appconfig.LoadedConfig
	defer func() { appconfig.LoadedConfig = originalConfig }()

	// Init the OauthConfig with AuthEnabled set to true
	appconfig.LoadedConfig = &appconfig.AppConfig{
		OauthConfig: &appconfig.OauthConfig{
			AuthEnabled:    true,
			IssuerUrl:      "https://issuer.example.com",
			RequiredScopes: "",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	hasBeenCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasBeenCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := OAuthTokenVerificationMiddleware(handler)
	middleware.ServeHTTP(rec, req)

	if hasBeenCalled {
		t.Fatal("expected handler to be called when auth is enabled")
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected an unauthorized status 401, got %v", rec.Code)
	}
}
