package middleware

import (
	"net/http"

	"github.com/ismailtemuroglu/discord/internal/platform/auth"
	"github.com/ismailtemuroglu/discord/internal/platform/httpx"
)

// CSRF validates the double-submit CSRF token on mutating requests.
//
// Safe methods (GET, HEAD, OPTIONS, TRACE) bypass the check.
// Paths listed in except bypass the check (auth bootstrap: register/login).
// Mutating methods require X-CSRF-Token to match the csrf_token cookie.
// Invalid or missing tokens return 403 Forbidden immediately.
func CSRF(except ...string) func(http.Handler) http.Handler {
	exempt := make(map[string]struct{}, len(except))
	for _, p := range except {
		exempt[p] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if auth.IsSafeMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			if _, ok := exempt[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			headerToken := r.Header.Get(auth.CSRFHeaderName)
			cookie, err := r.Cookie(auth.CSRFCookieName)
			if err != nil || !auth.TokensMatch(headerToken, cookie.Value) {
				_ = httpx.WriteError(w, &httpx.APIError{
					Code:    "CSRF_INVALID",
					Message: "CSRF token missing or invalid",
					Status:  http.StatusForbidden,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
