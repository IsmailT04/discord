package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ismailtemuroglu/discord/internal/identity/domain"
	"github.com/ismailtemuroglu/discord/internal/identity/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository is a Postgres-backed UserRepository.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository constructs a Postgres user repository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user. On success, user.ID/CreatedAt/UpdatedAt are set.
func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const q = `
		INSERT INTO users (email, username, password_hash, display_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.pool.QueryRow(ctx, q,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.DisplayName,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return mapWriteError(err)
	}
	return nil
}

// FindByID returns the user with the given id.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return r.findOne(ctx, `SELECT id, email, username, password_hash, display_name, created_at, updated_at
		FROM users WHERE id = $1`, id)
}

// FindByEmail returns the user with the given email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.findOne(ctx, `SELECT id, email, username, password_hash, display_name, created_at, updated_at
		FROM users WHERE email = $1`, email)
}

// FindByUsername returns the user with the given username.
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	return r.findOne(ctx, `SELECT id, email, username, password_hash, display_name, created_at, updated_at
		FROM users WHERE username = $1`, username)
}

func (r *UserRepository) findOne(ctx context.Context, q string, arg any) (*domain.User, error) {
	var u domain.User
	var displayName *string
	err := r.pool.QueryRow(ctx, q, arg).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.PasswordHash,
		&displayName,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	if displayName != nil {
		u.DisplayName = *displayName
	}
	return &u, nil
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch {
		case strings.Contains(pgErr.ConstraintName, "email"):
			return domain.ErrEmailTaken
		case strings.Contains(pgErr.ConstraintName, "username"):
			return domain.ErrUsernameTaken
		default:
			return domain.ErrEmailTaken
		}
	}
	return fmt.Errorf("insert user: %w", err)
}

var _ ports.UserRepository = (*UserRepository)(nil)
