package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

// setup struct for individual testCase
type testCase struct {
	name         string
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
	// save & defer restore the orig config
	originalConfig := appconfig.LoadedConfig
	defer func() { appconfig.LoadedConfig = originalConfig }()

	// setup handler for middleware
	hasBeenCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hasBeenCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// create a test server with the middleware
	middleware := OAuthTokenVerificationMiddleware(handler)
	ts := httptest.NewServer(middleware)
	defer ts.Close()

	// test cases
	testCases := []testCase{
		{
			name:         "Auth Disabled",
			authEnabled:  false,
			authHeader:   "",
			expectStatus: http.StatusOK,
			expectMesg:   "",
			expectNext:   true,
		},
		{
			name:         "Missing Authorization Header",
			authEnabled:  true,
			authHeader:   "",
			expectStatus: http.StatusUnauthorized,
			expectMesg:   "Authorization header missing\n",
			expectNext:   false,
		},
		{
			name:         "Invalid Authorization Header Format",
			authEnabled:  true,
			authHeader:   "Bearer", // current checks for <len("Bearer ")
			expectStatus: http.StatusUnauthorized,
			expectMesg:   "Authorization header format is invalid\n",
			expectNext:   false,
		},
		//{
		//	name:         "Valid JWT Token",
		//	authEnabled:  truecould not import github.com/golang-jwt/jwt/v4 (no required module provides package "github.com/golang-jwt/jwt/v4,
		//	authHeader:   "Bearer valid.jwt.token",
		//	expectStatus: http.StatusOK,
		//	expectMesg:   "",
		//	expectNext:   true,
		//},
	}

	// run the test cases
	for _, tc := range testCases {
		runOAuthTokenVerificationTestCase(t, ts, middleware, tc, &hasBeenCalled)
	}
}

// test case function
func runOAuthTokenVerificationTestCase(t *testing.T, ts *httptest.Server, middleware http.Handler, tc testCase, hasBeenCalled *bool) {
	t.Run(tc.name, func(t *testing.T) {
		// Reset the flag
		*hasBeenCalled = false

		// Mock the configuration
		appconfig.LoadedConfig = &appconfig.AppConfig{
			OauthConfig: &appconfig.OauthConfig{
				AuthEnabled: tc.authEnabled,
				IssuerUrl:   "http://example.com/oauth2",
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
