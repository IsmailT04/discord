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
	// RefreshTokenCookie is the long-lived rotating refresh cookie.
	RefreshTokenCookie = "refresh_token"
	// SessionIDCookie is an alternate session cookie name (legacy/compat).
	SessionIDCookie = "session_id"
)

// CookieOptions controls shared cookie attributes for auth cookies.
type CookieOptions struct {
	Domain string
	Secure bool
}

// SessionTokenFromRequest returns the opaque access session token from the request.
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

// RefreshTokenFromRequest returns the refresh_token cookie value, or "".
func RefreshTokenFromRequest(r *http.Request) string {
	if c, err := r.Cookie(RefreshTokenCookie); err == nil && c.Value != "" {
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

func setCookie(w http.ResponseWriter, name, value string, ttl time.Duration, httpOnly bool, opts CookieOptions) {
	maxAge := int(ttl.Seconds())
	expires := time.Now().Add(ttl)
	if ttl < 0 {
		maxAge = -1
		expires = time.Unix(0, 0)
		value = ""
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Domain:   cookieDomain(opts),
		MaxAge:   maxAge,
		Expires:  expires,
		HttpOnly: httpOnly,
		Secure:   opts.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetAccessToken sets the HttpOnly access session cookie.
func SetAccessToken(w http.ResponseWriter, token string, ttl time.Duration, opts CookieOptions) {
	setCookie(w, AccessTokenCookie, token, ttl, true, opts)
}

// ClearAccessToken removes the access session cookie.
func ClearAccessToken(w http.ResponseWriter, opts CookieOptions) {
	setCookie(w, AccessTokenCookie, "", -1, true, opts)
}

// SetRefreshToken sets the HttpOnly refresh cookie.
func SetRefreshToken(w http.ResponseWriter, token string, ttl time.Duration, opts CookieOptions) {
	setCookie(w, RefreshTokenCookie, token, ttl, true, opts)
}

// ClearRefreshToken removes the refresh cookie.
func ClearRefreshToken(w http.ResponseWriter, opts CookieOptions) {
	setCookie(w, RefreshTokenCookie, "", -1, true, opts)
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
	setCookie(w, CSRFCookieName, token, ttl, false, opts)
}

// ClearCSRFToken removes the CSRF cookie.
func ClearCSRFToken(w http.ResponseWriter, opts CookieOptions) {
	setCookie(w, CSRFCookieName, "", -1, false, opts)
}

// ClearAuthCookies clears access, refresh, and CSRF cookies.
func ClearAuthCookies(w http.ResponseWriter, opts CookieOptions) {
	ClearAccessToken(w, opts)
	ClearRefreshToken(w, opts)
	ClearCSRFToken(w, opts)
}
