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
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
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

// AuthResult is returned after successful register/login/refresh.
type AuthResult struct {
	User         domain.PublicProfile
	AccessToken  string
	RefreshToken string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
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

// Refresh rotates the refresh token and issues a new access + refresh pair.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*AuthResult, error) {
	if refreshToken == "" {
		return nil, domain.ErrInvalidRefresh
	}

	accessTTL, refreshTTL := s.ttls()
	pair, data, err := s.sessions.RotateRefresh(ctx, refreshToken, accessTTL, refreshTTL)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByID(ctx, data.UserID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user.ToPublic(),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		AccessTTL:    accessTTL,
		RefreshTTL:   refreshTTL,
	}, nil
}

// LogoutInput carries tokens to revoke for the current login family.
type LogoutInput struct {
	AccessToken  string
	RefreshToken string
}

// Logout revokes the current session family (access + refresh).
func (s *Service) Logout(ctx context.Context, in LogoutInput) error {
	var first error
	if in.AccessToken != "" {
		if err := s.sessions.Revoke(ctx, in.AccessToken); err != nil && first == nil {
			first = err
		}
	}
	if in.RefreshToken != "" {
		if err := s.sessions.RevokeByRefresh(ctx, in.RefreshToken); err != nil && first == nil {
			first = err
		}
	}
	return first
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
	accessTTL, refreshTTL := s.ttls()

	pair, err := s.sessions.Create(ctx, ports.SessionData{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, accessTTL, refreshTTL)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &AuthResult{
		User:         user.ToPublic(),
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		AccessTTL:    accessTTL,
		RefreshTTL:   refreshTTL,
	}, nil
}

func (s *Service) ttls() (access, refresh time.Duration) {
	access = s.cfg.AccessTokenTTL
	if access <= 0 {
		access = 15 * time.Minute
	}
	refresh = s.cfg.RefreshTokenTTL
	if refresh <= 0 {
		refresh = 7 * 24 * time.Hour
	}
	return access, refresh
}
