package auth

import (
	"context"
)

type userContextKey struct{}
type sessionIDContextKey struct{}

// User is the authenticated principal attached to a request context.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// SessionLookup resolves an opaque session token to a user.
// Returns (nil, nil) when the session is missing or expired.
type SessionLookup interface {
	Lookup(ctx context.Context, token string) (*User, error)
}

// WithUser stores the authenticated user in ctx.
func WithUser(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey{}, user)
}

// UserFromContext returns the authenticated user, or nil if absent.
func UserFromContext(ctx context.Context) *User {
	if ctx == nil {
		return nil
	}
	user, _ := ctx.Value(userContextKey{}).(*User)
	return user
}

// WithSessionID stores the session id associated with the request.
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDContextKey{}, sessionID)
}

// SessionIDFromContext returns the session id, or "" if absent.
func SessionIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(sessionIDContextKey{}).(string)
	return id
}
