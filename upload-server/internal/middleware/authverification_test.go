package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cdcgov/data-exchange-upload/upload-server/internal/appconfig"
)

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
	testCases := []struct {
		name         string
		authEnabled  bool
		authHeader   string
		expectStatus int
		expectMesg   string
		expectNext   bool
	}{
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
		//	authEnabled:  true,
		//	authHeader:   "Bearer valid.jwt.token",
		//	expectStatus: http.StatusOK,
		//	expectMesg:   "",
		//	expectNext:   true,
		//},
	}

	// run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// reset the flag
			hasBeenCalled = false

			// mock the configuration
			appconfig.LoadedConfig = &appconfig.AppConfig{
				OauthConfig: &appconfig.OauthConfig{
					AuthEnabled: tc.authEnabled,
					IssuerUrl:   "http://example.com/oauth2",
				},
			}

			// create a new request
			req := httptest.NewRequest(http.MethodGet, ts.URL, nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}

			// record the response
			rec := httptest.NewRecorder()

			// serve the request using the middleware
			middleware.ServeHTTP(rec, req)

			// check the status code
			if rec.Code != tc.expectStatus {
				t.Errorf("expected status %d, got %d", tc.expectStatus, rec.Code)
			}

			// check the body for status message
			if rec.Body.String() != tc.expectMesg {
				t.Errorf("expected message %q, got %q", tc.expectMesg, rec.Body.String())
			}

			// check if the next handler was called
			if hasBeenCalled != tc.expectNext {
				t.Errorf("expected next handler to be called: %v, got: %v", tc.expectNext, hasBeenCalled)
			}
		})
	}
}
