package middleware

import (
	"net/http"

	"github.com/ismailtemuroglu/discord/internal/platform/auth"
	"github.com/ismailtemuroglu/discord/internal/platform/httpx"
)

// CSRF validates the double-submit CSRF token on mutating requests.
//
// Safe methods (GET, HEAD, OPTIONS, TRACE) bypass the check.
// Mutating methods require X-CSRF-Token to match the csrf_token cookie.
// Invalid or missing tokens return 403 Forbidden immediately.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth.IsSafeMethod(r.Method) {
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
