package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

const (
	// AccessTokenCookie is the short-lived opaque access session cookie.
	AccessTokenCookie = "access_token"
	// SessionIDCookie is an alternate session cookie name (legacy/compat).
	SessionIDCookie = "session_id"
)

// CookieOptions controls shared cookie attributes for auth cookies.
type CookieOptions struct {
	Domain string
	Secure bool
	// MaxAge overrides TTL-derived MaxAge when clearing (use -1 / 0).
}

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

// cookieDomain returns opts.Domain, or empty for host-only cookies on localhost.
func cookieDomain(opts CookieOptions) string {
	if opts.Domain == "" || opts.Domain == "localhost" {
		return ""
	}
	return opts.Domain
}

// SetAccessToken sets the HttpOnly access session cookie.
func SetAccessToken(w http.ResponseWriter, token string, ttl time.Duration, opts CookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    token,
		Path:     "/",
		Domain:   cookieDomain(opts),
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
		HttpOnly: true,
		Secure:   opts.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAccessToken removes the access session cookie.
func ClearAccessToken(w http.ResponseWriter, opts CookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     "/",
		Domain:   cookieDomain(opts),
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   opts.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// NewCSRFToken generates a random CSRF token suitable for the double-submit cookie.
func NewCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// SetCSRFToken sets the non-HttpOnly CSRF cookie (readable by JS for double-submit).
func SetCSRFToken(w http.ResponseWriter, token string, ttl time.Duration, opts CookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    token,
		Path:     "/",
		Domain:   cookieDomain(opts),
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
		HttpOnly: false,
		Secure:   opts.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearCSRFToken removes the CSRF cookie.
func ClearCSRFToken(w http.ResponseWriter, opts CookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    "",
		Path:     "/",
		Domain:   cookieDomain(opts),
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: false,
		Secure:   opts.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthCookies clears access + CSRF cookies.
func ClearAuthCookies(w http.ResponseWriter, opts CookieOptions) {
	ClearAccessToken(w, opts)
	ClearCSRFToken(w, opts)
}
