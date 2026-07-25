package ports

import (
	"context"
	"time"

	"github.com/ismailtemuroglu/discord/internal/identity/domain"
)

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
}

// SessionData is the opaque session payload stored in the session backend.
type SessionData struct {
	UserID   string
	Username string
	Email    string
}

// TokenPair is an access + refresh token issued together.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// SessionStore creates, rotates, and revokes opaque access + refresh sessions.
// Lookup for middleware is provided by the same adapter via auth.SessionLookup.
type SessionStore interface {
	// Create issues a new access + refresh pair under a fresh token family.
	Create(ctx context.Context, data SessionData, accessTTL, refreshTTL time.Duration) (TokenPair, error)
	// RotateRefresh consumes refreshToken and issues a new pair (same family).
	// Detects reuse of already-rotated tokens and revokes the family.
	RotateRefresh(ctx context.Context, refreshToken string, accessTTL, refreshTTL time.Duration) (TokenPair, SessionData, error)
	// Revoke deletes the access session.
	Revoke(ctx context.Context, accessToken string) error
	// RevokeByRefresh revokes the refresh token's family (access + all refresh tokens).
	RevokeByRefresh(ctx context.Context, refreshToken string) error
}

// PasswordHasher hashes and verifies passwords (argon2id).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}
