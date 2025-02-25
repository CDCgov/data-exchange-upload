package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// global private and public keys for JWT signature and validation testing
var privateKey *rsa.PrivateKey
var publicKey rsa.PublicKey

// setup struct for individual test case
type testCase struct {
	name                     string
	issuerURL                string
	route                    string
	authEnabled              bool // flag to test when auth enabled/disabled
	authHeader               string
	sessionCookie            *http.Cookie
	expectStatus             int    // expected HTTP status code in response
	expectMesg               string // expected error response body message
	expectNext               bool   // false = has error response in middleware, true = passes on to next handler
	expectedRedirectLocation string
	requiredScopes           string // "" for no required scopes
}

// tests the VerifyOAuthTokenMiddleware for multiple cases
func TestVerifyOAuthTokenMiddleware_TestCases(t *testing.T) {
	// init RSA keys for signing and verification
	err := initKeys()
	if err != nil {
		t.Fatalf("failed to initialize keys: %v", err)
	}

	// start the mock OIDC server
	mockOIDC := mockOIDCServer()
	defer mockOIDC.Close()

	// get the dynamic issuer url
	issuerURL := mockOIDC.URL

	// create VALID mock token w/ +1-hour expire offset
	mockTokenValid, _ := createMockJWT(issuerURL, 1, "")
	mockValidSessionCookie := &http.Cookie{
		Name:  UserSessionCookieName,
		Value: mockTokenValid,
	}
	// create mock token by concat a Z to make an invalid signature
	mockTokenInvalidSignature := mockTokenValid + "Z"

	// create EXPIRED mock token w/ -1-hour expire offset
	mockTokenExpired, _ := createMockJWT(issuerURL, -1, "")

	// create token with wrong issuer
	mockTokenWrongIssuer, _ := createMockJWT("http://wrong-issuer.com", 1, "")

	// create VALID mock token w/ +1-hour expire offset with scope
	mockTokenValidWithScope, _ := createMockJWT(issuerURL, 1, "testscope1 testscope2")

	// setup up VALID mock token w/ +1-hour expire offset that includes the scopes
	mockTokenValidIncludesReqScopes, _ := createMockJWT(issuerURL, 1, "read:scope1 read:custom1 write:scope1 write:custom1")

	// test cases list
	testCases := []testCase{
		{
			name:           "Auth Disabled",
			issuerURL:      issuerURL,
			authEnabled:    false,
			authHeader:     "",
			expectStatus:   http.StatusOK,
			expectMesg:     "",
			expectNext:     true,
			requiredScopes: "",
		},
		{
			name:           "No Token Provided In Request",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "",
			expectStatus:   http.StatusUnauthorized,
			expectMesg:     "authorization token not found",
			expectNext:     false,
			requiredScopes: "",
		},
		{
			name:           "Invalid Authorization Header Format",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer", // related code checks for <len("Bearer ")
			expectStatus:   http.StatusUnauthorized,
			expectMesg:     "authorization header format is invalid",
			expectNext:     false,
			requiredScopes: "",
		},
		{
			name:           "Expired JWT Token",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer " + mockTokenExpired,
			expectStatus:   http.StatusUnauthorized,
			expectMesg:     "failed to verify token\noidc: token is expired",
			expectNext:     false,
			requiredScopes: "",
		},
		{
			name:           "Invalid JWT Signature",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer " + mockTokenInvalidSignature,
			expectStatus:   http.StatusUnauthorized,
			expectMesg:     "failed to verify token\nfailed to verify signature:",
			expectNext:     false,
			requiredScopes: "",
		},
		{
			name:           "Invalid Issuer",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer " + mockTokenWrongIssuer,
			expectStatus:   http.StatusUnauthorized,
			expectMesg:     "failed to verify token\noidc: id token issued by a different provider",
			expectNext:     false,
			requiredScopes: "",
		},
		{
			name:           "Valid JWT Has Scope Claim Server Does Not Require",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer " + mockTokenValidWithScope,
			expectStatus:   http.StatusOK,
			expectMesg:     "",
			expectNext:     true,
			requiredScopes: "",
		},
		{
			name:           "Valid JWT Token In Auth Header",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer " + mockTokenValid,
			expectStatus:   http.StatusOK,
			expectMesg:     "",
			expectNext:     true,
			requiredScopes: "",
		},
		{
			name:           "Valid JWT Token In Cookie",
			issuerURL:      issuerURL,
			authEnabled:    true,
			expectStatus:   http.StatusOK,
			expectMesg:     "",
			expectNext:     true,
			requiredScopes: "",
			sessionCookie:  mockValidSessionCookie,
		},
		// RequiredScopes related tests
		{
			name:           "JWT with no scope claim",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer " + mockTokenValid,
			expectStatus:   http.StatusForbidden,
			expectMesg:     "one or more required scopes not found",
			expectNext:     false,
			requiredScopes: "read:scope1",
		},
		{
			name:           "JWT Token includes custom, 2 required scopes, missing one req scope",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer " + mockTokenValidIncludesReqScopes,
			expectStatus:   http.StatusForbidden,
			expectMesg:     "one or more required scopes not found",
			expectNext:     false,
			requiredScopes: "read:scope1 write:scope1 read:scope2",
		},
		{
			name:           "Valid JWT Token includes custom and both required scopes",
			issuerURL:      issuerURL,
			authEnabled:    true,
			authHeader:     "Bearer " + mockTokenValidIncludesReqScopes,
			expectStatus:   http.StatusOK,
			expectMesg:     "",
			expectNext:     true,
			requiredScopes: "read:scope1 write:scope1",
		},
	}

	// run the test cases
	for _, tc := range testCases {
		runOAuthTokenVerificationTestCase(t, tc)
	}
}

// test case function
func runOAuthTokenVerificationTestCase(t *testing.T, tc testCase) {

	t.Run(tc.name, func(t *testing.T) {
		// create handler for middleware
		hasBeenCalled := false
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hasBeenCalled = true
			w.WriteHeader(http.StatusOK)
		})

		middleware, err := NewAuthMiddleware(appconfig.OauthConfig{
			AuthEnabled:    tc.authEnabled,
			IssuerUrl:      tc.issuerURL,
			RequiredScopes: tc.requiredScopes,
		})
		if err != nil {
			t.Fatal(err)
		}

		// create a test server with the middleware
		middlewareHandler := middleware.VerifyOAuthTokenMiddleware(handler)
		ts := httptest.NewServer(middlewareHandler)
		defer ts.Close()

		// create a new request
		req := httptest.NewRequest(http.MethodGet, ts.URL, nil)
		if tc.authHeader != "" {
			req.Header.Set("Authorization", tc.authHeader)
		}
		if tc.sessionCookie != nil {
			req.AddCookie(tc.sessionCookie)
		}

		// record the response
		rec := httptest.NewRecorder()

		// serve the request using the middleware
		middlewareHandler.ServeHTTP(rec, req)

		// check the status code
		if rec.Code != tc.expectStatus {
			t.Errorf("expected status %d, got %d", tc.expectStatus, rec.Code)
		}

		// check the body for status message
		if !strings.HasPrefix(rec.Body.String(), tc.expectMesg) {
			t.Errorf("expected message %q, got %q", tc.expectMesg, rec.Body.String())
		}

		// check if the next handler was called
		if hasBeenCalled != tc.expectNext {
			t.Errorf("expected next handler to be called: %v, got: %v", tc.expectNext, hasBeenCalled)
		}
	})
}

func TestUserSessionMiddleware_TestCases(t *testing.T) {
	err := initKeys()
	if err != nil {
		t.Fatalf("failed to initialize keys: %v", err)
	}

	mockOIDC := mockOIDCServer()
	defer mockOIDC.Close()
	issuerURL := mockOIDC.URL
	mockTokenValid, _ := createMockJWT(issuerURL, 1, "")
	mockValidSessionCookie := &http.Cookie{
		Name:  UserSessionCookieName,
		Value: mockTokenValid,
	}
	mockTokenExpired, _ := createMockJWT(issuerURL, -1, "")
	mockSessionWithExpiredToken := &http.Cookie{
		Name:  UserSessionCookieName,
		Value: mockTokenExpired,
	}
	mockSessionWithBadCookieName := &http.Cookie{
		Name:  "bogus",
		Value: mockTokenValid,
	}

	testCases := []testCase{
		{
			name:         "Auth Disabled",
			issuerURL:    issuerURL,
			authEnabled:  false,
			expectStatus: http.StatusOK,
			expectMesg:   "",
			expectNext:   true,
		},
		{
			name:          "Valid Session Cookie",
			issuerURL:     issuerURL,
			authEnabled:   true,
			sessionCookie: mockValidSessionCookie,
			expectStatus:  http.StatusOK,
			expectMesg:    "",
			expectNext:    true,
		},
		{
			name:                     "Missing Session Cookie",
			issuerURL:                issuerURL,
			authEnabled:              true,
			expectStatus:             http.StatusSeeOther,
			expectNext:               false,
			expectedRedirectLocation: "/login",
		},
		{
			name:                     "Expired Token",
			issuerURL:                issuerURL,
			authEnabled:              true,
			sessionCookie:            mockSessionWithExpiredToken,
			expectStatus:             http.StatusSeeOther,
			expectNext:               false,
			expectedRedirectLocation: "/login",
		},
		{
			name:                     "Invalid Cookie Name",
			issuerURL:                issuerURL,
			authEnabled:              true,
			sessionCookie:            mockSessionWithBadCookieName,
			expectStatus:             http.StatusSeeOther,
			expectNext:               false,
			expectedRedirectLocation: "/login",
		},
		{
			name:                     "Redirect to Other Page",
			issuerURL:                issuerURL,
			authEnabled:              true,
			sessionCookie:            mockSessionWithExpiredToken,
			expectStatus:             http.StatusSeeOther,
			expectMesg:               "",
			expectNext:               false,
			route:                    "/status/1234",
			expectedRedirectLocation: "/login?redirect=/status/1234",
		},
		{
			name:                     "Redirect with Query Params",
			issuerURL:                issuerURL,
			authEnabled:              true,
			sessionCookie:            mockSessionWithExpiredToken,
			expectStatus:             http.StatusSeeOther,
			expectMesg:               "",
			expectNext:               false,
			route:                    "/manifest?data_stream=test&data_stream_route=test",
			expectedRedirectLocation: "/login?redirect=/manifest?data_stream=test&data_stream_route=test",
		},
	}

	for _, tc := range testCases {
		runUserSessionMiddlewareTestCase(t, tc)
	}
}

func runUserSessionMiddlewareTestCase(t *testing.T, tc testCase) {
	t.Run(tc.name, func(t *testing.T) {
		hasBeenCalled := false
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hasBeenCalled = true
			w.WriteHeader(http.StatusOK)
		})

		middleware, err := NewAuthMiddleware(appconfig.OauthConfig{
			AuthEnabled:    tc.authEnabled,
			IssuerUrl:      tc.issuerURL,
			RequiredScopes: tc.requiredScopes,
		})
		if err != nil {
			t.Fatal(err)
		}
		handler := middleware.VerifyUserSession(nextHandler)
		ts := httptest.NewServer(handler)
		defer ts.Close()

		req := httptest.NewRequest(http.MethodGet, ts.URL+tc.route, nil)
		if tc.sessionCookie != nil {
			req.AddCookie(tc.sessionCookie)
		}
		resp := httptest.NewRecorder()
		handler.ServeHTTP(resp, req)

		if resp.Code != tc.expectStatus {
			t.Errorf("expected status %d, got %d", tc.expectStatus, resp.Code)
		}

		if hasBeenCalled != tc.expectNext {
			t.Errorf("expected next handler to be called: %v, got: %v", tc.expectNext, hasBeenCalled)
		}

		if tc.expectedRedirectLocation != "" {
			redirectUrl, err := resp.Result().Location()
			if err != nil {
				t.Error(err)
			}
			decoded, err := url.QueryUnescape(redirectUrl.String())
			if err != nil {
				t.Error(err)
			}
			if tc.expectedRedirectLocation != decoded {
				t.Errorf("expected redirect to %s, got %s", tc.expectedRedirectLocation, decoded)
			}
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
func createMockJWT(issuerURL string, expireOffset time.Duration, scopes string) (string, error) {
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

	// add the "scope" claim if scopes is not empty
	if scopes != "" {
		claims["scope"] = scopes
	}

	// create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// sign the token w/ private key
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
