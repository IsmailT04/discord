package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ismailtemuroglu/discord/internal/identity/domain"
	"github.com/ismailtemuroglu/discord/internal/identity/ports"
)

// Config holds TTLs and other application-level auth settings.
type Config struct {
	AccessTokenTTL time.Duration
}

// Service orchestrates identity use cases.
type Service struct {
	users    ports.UserRepository
	sessions ports.SessionStore
	hasher   ports.PasswordHasher
	cfg      Config
}

// New constructs an identity application service.
func New(
	users ports.UserRepository,
	sessions ports.SessionStore,
	hasher ports.PasswordHasher,
	cfg Config,
) *Service {
	return &Service{
		users:    users,
		sessions: sessions,
		hasher:   hasher,
		cfg:      cfg,
	}
}

// RegisterInput is the register use-case input.
type RegisterInput struct {
	Email       string
	Username    string
	Password    string
	DisplayName string
}

// AuthResult is returned after successful register/login.
type AuthResult struct {
	User         domain.PublicProfile
	AccessToken  string
	AccessTTL    time.Duration
}

// Register creates a user and opens a session.
func (s *Service) Register(ctx context.Context, in RegisterInput) (*AuthResult, error) {
	if err := domain.ValidatePassword(in.Password); err != nil {
		return nil, err
	}

	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := domain.NewUser(in.Email, in.Username, hash, in.DisplayName)
	if err != nil {
		return nil, err
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.openSession(ctx, user)
}

// LoginInput is the login use-case input.
type LoginInput struct {
	Email    string
	Password string
}

// Login authenticates by email/password and opens a session.
func (s *Service) Login(ctx context.Context, in LoginInput) (*AuthResult, error) {
	email, err := domain.NormalizeEmail(in.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if in.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := s.hasher.Compare(user.PasswordHash, in.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return s.openSession(ctx, user)
}

// Logout revokes the given session token.
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	return s.sessions.Revoke(ctx, token)
}

// Me returns the public profile for userID.
func (s *Service) Me(ctx context.Context, userID string) (*domain.PublicProfile, error) {
	if userID == "" {
		return nil, domain.ErrUnauthenticated
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile := user.ToPublic()
	return &profile, nil
}

func (s *Service) openSession(ctx context.Context, user *domain.User) (*AuthResult, error) {
	ttl := s.cfg.AccessTokenTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	token, err := s.sessions.Create(ctx, ports.SessionData{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, ttl)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthResult{
		User:        user.ToPublic(),
		AccessToken: token,
		AccessTTL:   ttl,
	}, nil
}
