package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

// mock the oidc conf response
func mockOIDCServer() *httptest.Server {
	mux := http.NewServeMux()

	// init new test server
	testServer := httptest.NewServer(mux)

	// get the dynamic url after start
	issuer := testServer.URL

	// mock the oidc discovery document
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"issuer":                 issuer,
			"authorization_endpoint": issuer + "/oauth2/authorize",
			"token_endpoint":         issuer + "/oauth2/token",
			"jwks_uri":               issuer + "/oauth2/jwks",
		}
		json.NewEncoder(w).Encode(config)
	})

	// mock jwks endpoint needed for verification
	mux.HandleFunc("/oauth2/jwks", func(w http.ResponseWriter, r *http.Request) {
		keys := map[string]interface{}{
			"keys": []map[string]string{
				{
					"kty": "RSA",
					"kid": "test-key-id",
					"n":   "test-modulus",
					"e":   "AQAB",
				},
			},
		}
		json.NewEncoder(w).Encode(keys)
	})

	return testServer
}

// helper to create a mock jwt token
func createMockJWT(issuerURL string, expireOffset time.Duration) string {
	// set the expiration time for offset (use negative to test expire)
	expirationTime := time.Now().Add(expireOffset * time.Hour).Unix()

	// create a simple jwt
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"sub":"1234567890","name":"John Doe","iat":1516239022,"iss":"%s","exp":%d}`, issuerURL, expirationTime)))
	signature := base64.RawURLEncoding.EncodeToString([]byte("signature"))

	// return in the format: header.payload.signature
	return header + "." + payload + "." + signature
}

func TestOAuthTokenVerificationMiddleware_WithMockOIDC(t *testing.T) {
	// start the mock server
	mockOIDC := mockOIDCServer()
	defer mockOIDC.Close()

	// get the dynamic issuer url
	issuerURL := mockOIDC.URL

	fmt.Printf("\n\n   issuerURL: %s", issuerURL)
	fmt.Printf("\n\n   mockOIDC.URL: %s", mockOIDC.URL)

	// mock the AppConfig with issuer
	appconfig.LoadedConfig = &appconfig.AppConfig{
		OauthConfig: &appconfig.OauthConfig{
			AuthEnabled: true,
			IssuerUrl:   issuerURL, // use mock server url
		},
	}

	// setup up mock token w/ +1-hour expire offset
	//mockToken := createMockJWT(issuerURL, -1) // expired test
	mockToken := createMockJWT(issuerURL, 1)

	// set up the handler
	hasBeenCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasBeenCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// create middleware with handler
	middleware := OAuthTokenVerificationMiddleware(handler)

	// create test server with the middleware
	ts := httptest.NewServer(middleware)
	defer ts.Close()

	// create a request with a valid jwt token
	req := httptest.NewRequest(http.MethodGet, ts.URL, nil)
	req.Header.Set("Authorization", "Bearer "+mockToken)

	// record the response
	rec := httptest.NewRecorder()

	// serve the request using the middleware
	middleware.ServeHTTP(rec, req)

	fmt.Printf("\n\n  ----   rec %s", rec.Body)

	// check the status code
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// check if the next handler was called
	if !hasBeenCalled {
		t.Errorf("expected next handler to be called")
	}
}
