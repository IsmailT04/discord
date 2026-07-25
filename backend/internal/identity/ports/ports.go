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

// SessionStore creates and revokes opaque sessions (lookup is also used by
// platform middleware via auth.SessionLookup on the same adapter).
type SessionStore interface {
	Create(ctx context.Context, data SessionData, ttl time.Duration) (token string, err error)
	Revoke(ctx context.Context, token string) error
}

// PasswordHasher hashes and verifies passwords (argon2id).
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}
