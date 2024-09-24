package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"github.com/golang-jwt/jwt/v5"
)

// global private and public keys
var privateKey *rsa.PrivateKey
var publicKey rsa.PublicKey

// setup struct for individual testCase
type testCase struct {
	name         string
	issuerURL    string
	authEnabled  bool
	authHeader   string
	expectStatus int
	expectMesg   string
	expectNext   bool
}

// TestOAuthTokenVerificationMiddleware_TestCases
//
//	tests the OAuthTokenVerificationMiddleware_TestCases for the following cases:
//	  - auth is disabled
//	  - missing auth header
//	  - invalid auth header format - TODO
//	  - valid jwt token - TODO
//	  - valid opaque token - TODO
func TestOAuthTokenVerificationMiddleware_TestCases(t *testing.T) {
	// init RSA keys for signing and verification
	err := initKeys()
	if err != nil {
		t.Fatalf("failed to initialize keys: %v", err)
	}

	// start the mock server
	mockOIDC := mockOIDCServer()
	defer mockOIDC.Close()

	// get the dynamic issuer url
	issuerURL := mockOIDC.URL

	fmt.Printf("\n\n   issuerURL: %s", issuerURL)
	fmt.Printf("\n\n   mockOIDC.URL: %s", mockOIDC.URL)

	// create handler for middleware
	hasBeenCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasBeenCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// create a test server with the middleware
	middleware := OAuthTokenVerificationMiddleware(handler)
	ts := httptest.NewServer(middleware)
	defer ts.Close()

	// setup up VALID mock token w/ +1-hour expire offset
	mockTokenValid, err := createMockJWT(issuerURL, 1)
	if err != nil {
		t.Fatalf("failed to create mock jwt token: %v", err)
	}

	//fmt.Printf("\n\n   mockToken: %s", mockToken)

	// test cases
	testCases := []testCase{
		{
			name:         "Auth Disabled",
			issuerURL:    issuerURL,
			authEnabled:  false,
			authHeader:   "",
			expectStatus: http.StatusOK,
			expectMesg:   "",
			expectNext:   true,
		},
		{
			name:         "Missing Authorization Header",
			issuerURL:    issuerURL,
			authEnabled:  true,
			authHeader:   "",
			expectStatus: http.StatusUnauthorized,
			expectMesg:   "Authorization header missing\n",
			expectNext:   false,
		},
		{
			name:         "Invalid Authorization Header Format",
			issuerURL:    issuerURL,
			authEnabled:  true,
			authHeader:   "Bearer", // current checks for <len("Bearer ")
			expectStatus: http.StatusUnauthorized,
			expectMesg:   "Authorization header format is invalid\n",
			expectNext:   false,
		},
		{
			name:         "Valid JWT Token",
			issuerURL:    issuerURL,
			authEnabled:  true,
			authHeader:   "Bearer " + mockTokenValid,
			expectStatus: http.StatusOK,
			expectMesg:   "",
			expectNext:   true,
		},
	}

	// run the test cases
	for _, tc := range testCases {
		runOAuthTokenVerificationTestCase(t, ts, middleware, tc, &hasBeenCalled)
	}
}

// test case function
func runOAuthTokenVerificationTestCase(t *testing.T, ts *httptest.Server, middleware http.Handler, tc testCase, hasBeenCalled *bool) {

	t.Run(tc.name, func(t *testing.T) {
		// save & defer restore the orig config
		originalConfig := appconfig.LoadedConfig
		defer func() { appconfig.LoadedConfig = originalConfig }()

		// Reset the flag
		*hasBeenCalled = false

		// Mock the configuration
		appconfig.LoadedConfig = &appconfig.AppConfig{
			OauthConfig: &appconfig.OauthConfig{
				AuthEnabled: tc.authEnabled,
				IssuerUrl:   tc.issuerURL,
			},
		}

		// Create a new request
		req := httptest.NewRequest(http.MethodGet, ts.URL, nil)
		if tc.authHeader != "" {
			req.Header.Set("Authorization", tc.authHeader)
		}

		// Record the response
		rec := httptest.NewRecorder()

		// Serve the request using the middleware
		middleware.ServeHTTP(rec, req)

		// Check the status code
		if rec.Code != tc.expectStatus {
			t.Errorf("expected status %d, got %d", tc.expectStatus, rec.Code)
		}

		// Check the body for status message
		if rec.Body.String() != tc.expectMesg {
			t.Errorf("expected message %q, got %q", tc.expectMesg, rec.Body.String())
		}

		// Check if the next handler was called
		if *hasBeenCalled != tc.expectNext {
			t.Errorf("expected next handler to be called: %v, got: %v", tc.expectNext, *hasBeenCalled)
		}
	})
}

// initialize keys for testing
func initKeys() error {
	var err error
	// generate private key
	privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	// get the public key
	publicKey = privateKey.PublicKey
	return nil
}

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

	mux.HandleFunc("/oauth2/jwks", func(w http.ResponseWriter, r *http.Request) {
		key := map[string]interface{}{
			"kty": "RSA",
			"kid": "test-key-id",
			"n":   base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes()),                    // mod
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes()), // exp
		}
		keys := map[string]interface{}{
			"keys": []map[string]interface{}{key},
		}
		json.NewEncoder(w).Encode(keys)
	})

	return testServer
}

// helper to create a mock jwt token
func createMockJWT(issuerURL string, expireOffset time.Duration) (string, error) {
	// set the expiration time for offset (use negative to test expire)
	expirationTime := time.Now().Add(expireOffset * time.Hour).Unix()

	// create claims
	claims := jwt.MapClaims{
		"sub":  "1234567890",
		"name": "John Doe",
		"iat":  time.Now().Unix(),
		"iss":  issuerURL,
		"exp":  expirationTime,
	}

	// create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// sign the token w/ private key
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	// return signedToken + "z", nil // test bad signature
	return signedToken, nil
}
