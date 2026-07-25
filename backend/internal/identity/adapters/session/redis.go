package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ismailtemuroglu/discord/internal/identity/ports"
	"github.com/ismailtemuroglu/discord/internal/platform/auth"
	"github.com/redis/go-redis/v9"
)

const keyPrefix = "session:"

// redisSession is the JSON payload stored under session:{token}.
type redisSession struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
}

// RedisStore creates, looks up, and revokes opaque sessions in Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a session store backed by Redis.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// Create stores a new session and returns the opaque token.
func (s *RedisStore) Create(ctx context.Context, data ports.SessionData, ttl time.Duration) (string, error) {
	token, err := newToken()
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(redisSession{
		UserID:   data.UserID,
		Username: data.Username,
		Email:    data.Email,
	})
	if err != nil {
		return "", fmt.Errorf("redis session encode: %w", err)
	}

	if err := s.client.Set(ctx, keyPrefix+token, payload, ttl).Err(); err != nil {
		return "", fmt.Errorf("redis session set: %w", err)
	}
	return token, nil
}

// Lookup returns the user for token, or (nil, nil) if the session is missing.
func (s *RedisStore) Lookup(ctx context.Context, token string) (*auth.User, error) {
	if token == "" {
		return nil, nil
	}

	raw, err := s.client.Get(ctx, keyPrefix+token).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis session get: %w", err)
	}

	var sess redisSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return nil, fmt.Errorf("redis session decode: %w", err)
	}
	if sess.UserID == "" {
		return nil, nil
	}

	return &auth.User{
		ID:       sess.UserID,
		Username: sess.Username,
		Email:    sess.Email,
	}, nil
}

// Revoke deletes the session for token. Missing keys are ignored.
func (s *RedisStore) Revoke(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.client.Del(ctx, keyPrefix+token).Err(); err != nil {
		return fmt.Errorf("redis session del: %w", err)
	}
	return nil
}

func newToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

var (
	_ auth.SessionLookup  = (*RedisStore)(nil)
	_ ports.SessionStore  = (*RedisStore)(nil)
)
