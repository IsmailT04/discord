package domain

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Domain errors — map to HTTP in adapters, never leak status codes here.
var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidDisplayName = errors.New("invalid display name")
	ErrEmailTaken         = errors.New("email already registered")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrRefreshReuse       = errors.New("refresh token reuse detected")
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

const (
	MinPasswordLen = 8
	MaxPasswordLen = 128
	MaxDisplayName = 64
)

// User is the identity aggregate root.
type User struct {
	ID           string
	Email        string
	Username     string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser validates registration fields and builds a user with the given hash.
func NewUser(email, username, passwordHash, displayName string) (*User, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return nil, err
	}
	username, err = NormalizeUsername(username)
	if err != nil {
		return nil, err
	}
	if passwordHash == "" {
		return nil, ErrInvalidPassword
	}
	displayName, err = NormalizeDisplayName(displayName, username)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &User{
		Email:        email,
		Username:     username,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

// NormalizeEmail trims and lowercases email, then validates format.
func NormalizeEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || utf8.RuneCountInString(email) > 254 {
		return "", ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return "", ErrInvalidEmail
	}
	return email, nil
}

// NormalizeUsername trims and validates username rules.
func NormalizeUsername(username string) (string, error) {
	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return "", fmt.Errorf("%w: must be 3-32 chars of [a-zA-Z0-9_]", ErrInvalidUsername)
	}
	return username, nil
}

// NormalizeDisplayName uses username when empty.
func NormalizeDisplayName(displayName, username string) (string, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return username, nil
	}
	n := utf8.RuneCountInString(displayName)
	if n < 1 || n > MaxDisplayName {
		return "", ErrInvalidDisplayName
	}
	return displayName, nil
}

// ValidatePassword checks plaintext password policy (before hashing).
func ValidatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if n < MinPasswordLen || n > MaxPasswordLen {
		return fmt.Errorf("%w: must be %d-%d characters", ErrInvalidPassword, MinPasswordLen, MaxPasswordLen)
	}
	return nil
}

// PublicProfile is the safe projection returned to clients.
type PublicProfile struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToPublic returns a PublicProfile for u.
func (u *User) ToPublic() PublicProfile {
	return PublicProfile{
		ID:          u.ID,
		Email:       u.Email,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		CreatedAt:   u.CreatedAt,
	}
}
