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

func TestOAuthTokenVerificationMiddleware_WithMockOIDC(t *testing.T) {
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

	// mock the AppConfig with issuer
	appconfig.LoadedConfig = &appconfig.AppConfig{
		OauthConfig: &appconfig.OauthConfig{
			AuthEnabled: true,
			IssuerUrl:   issuerURL, // use mock server url
		},
	}

	// setup up mock token w/ +1-hour expire offset
	//mockToken := createMockJWT(issuerURL, -1) // expired test
	mockToken, err := createMockJWT(issuerURL, 1)
	if err != nil {
		t.Fatalf("failed to create mock jwt token: %v", err)
	}

	fmt.Printf("\n\n   mockToken: %s", mockToken)

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
