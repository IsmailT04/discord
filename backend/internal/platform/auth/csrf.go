package auth

import (
	"crypto/subtle"
	"net/http"
)

const (
	// CSRFCookieName is the double-submit CSRF cookie (readable by JS).
	CSRFCookieName = "csrf_token"
	// CSRFHeaderName is the header the SPA must echo on mutating requests.
	CSRFHeaderName = "X-CSRF-Token"
)

// SafeMethods are HTTP methods that must not mutate state and skip CSRF checks.
var SafeMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

// IsSafeMethod reports whether method is exempt from CSRF validation.
func IsSafeMethod(method string) bool {
	_, ok := SafeMethods[method]
	return ok
}

// TokensMatch reports whether header and cookie CSRF tokens are equal
// using a constant-time comparison.
func TokensMatch(headerToken, cookieToken string) bool {
	if headerToken == "" || cookieToken == "" {
		return false
	}
	if len(headerToken) != len(cookieToken) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(headerToken), []byte(cookieToken)) == 1
}
