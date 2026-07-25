package session

import (
	"context"
	"encoding/json"
	"fmt"

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

// RedisStore looks up opaque sessions in Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a SessionLookup backed by Redis.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
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

// Ensure RedisStore satisfies auth.SessionLookup.
var _ auth.SessionLookup = (*RedisStore)(nil)
