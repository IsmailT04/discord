package auth

import "net/http"

const (
	// AccessTokenCookie is the short-lived opaque access session cookie.
	AccessTokenCookie = "access_token"
	// SessionIDCookie is an alternate session cookie name (legacy/compat).
	SessionIDCookie = "session_id"
)

// SessionTokenFromRequest returns the opaque session token from the request.
// Prefers access_token, then falls back to session_id. Empty if neither is set.
func SessionTokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(AccessTokenCookie); err == nil && c.Value != "" {
		return c.Value
	}
	if c, err := r.Cookie(SessionIDCookie); err == nil && c.Value != "" {
		return c.Value
	}
	return ""
}
