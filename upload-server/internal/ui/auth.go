package ui

import (
	"log/slog"
	"net/http"
)

func ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Info("validating token")
		// validate token
		// if invalid, redirect to /login
		// if valid, set cookie if it's not set
		next.ServeHTTP(w, r)
	})
}
