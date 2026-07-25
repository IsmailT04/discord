package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/ismailtemuroglu/discord/internal/identity/domain"
	"github.com/ismailtemuroglu/discord/internal/identity/ports"
	"github.com/ismailtemuroglu/discord/internal/platform/auth"
	"github.com/redis/go-redis/v9"
)

const (
	sessionPrefix = "session:"
	refreshPrefix = "refresh:"
	familyPrefix  = "family:"
)

// redisSession is stored under session:{accessToken}.
type redisSession struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email,omitempty"`
	FamilyID string `json:"family_id"`
}

// redisRefresh is stored under refresh:{refreshToken}.
type redisRefresh struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Email       string `json:"email,omitempty"`
	FamilyID    string `json:"family_id"`
	AccessToken string `json:"access_token"`
	ReplacedBy  string `json:"replaced_by,omitempty"`
}

// RedisStore creates, looks up, rotates, and revokes opaque sessions in Redis.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a session store backed by Redis.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// Create issues a new access + refresh pair under a new family.
func (s *RedisStore) Create(ctx context.Context, data ports.SessionData, accessTTL, refreshTTL time.Duration) (ports.TokenPair, error) {
	var zero ports.TokenPair

	familyID, err := newToken()
	if err != nil {
		return zero, err
	}
	access, err := newToken()
	if err != nil {
		return zero, err
	}
	refresh, err := newToken()
	if err != nil {
		return zero, err
	}

	if err := s.writePair(ctx, data, familyID, access, refresh, accessTTL, refreshTTL); err != nil {
		return zero, err
	}
	return ports.TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

// Lookup returns the user for an access token, or (nil, nil) if missing.
func (s *RedisStore) Lookup(ctx context.Context, token string) (*auth.User, error) {
	if token == "" {
		return nil, nil
	}

	raw, err := s.client.Get(ctx, sessionPrefix+token).Bytes()
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

// RotateRefresh consumes a refresh token and issues a new pair (same family).
func (s *RedisStore) RotateRefresh(ctx context.Context, refreshToken string, accessTTL, refreshTTL time.Duration) (ports.TokenPair, ports.SessionData, error) {
	var zero ports.TokenPair
	var zeroData ports.SessionData

	if refreshToken == "" {
		return zero, zeroData, domain.ErrInvalidRefresh
	}

	key := refreshPrefix + refreshToken
	raw, err := s.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return zero, zeroData, domain.ErrInvalidRefresh
	}
	if err != nil {
		return zero, zeroData, fmt.Errorf("redis refresh get: %w", err)
	}

	var old redisRefresh
	if err := json.Unmarshal(raw, &old); err != nil {
		return zero, zeroData, fmt.Errorf("redis refresh decode: %w", err)
	}

	revoked, err := s.familyRevoked(ctx, old.FamilyID)
	if err != nil {
		return zero, zeroData, err
	}
	if revoked {
		return zero, zeroData, domain.ErrInvalidRefresh
	}

	// Reuse of an already-rotated token → revoke the whole family.
	if old.ReplacedBy != "" {
		_ = s.revokeFamily(ctx, old.FamilyID)
		return zero, zeroData, domain.ErrRefreshReuse
	}

	newAccess, err := newToken()
	if err != nil {
		return zero, zeroData, err
	}
	newRefresh, err := newToken()
	if err != nil {
		return zero, zeroData, err
	}

	data := ports.SessionData{
		UserID:   old.UserID,
		Username: old.Username,
		Email:    old.Email,
	}

	// Mark old refresh as rotated (kept until TTL for reuse detection).
	old.ReplacedBy = newRefresh
	marked, err := json.Marshal(old)
	if err != nil {
		return zero, zeroData, fmt.Errorf("redis refresh encode: %w", err)
	}
	ttl := s.client.TTL(ctx, key).Val()
	if ttl <= 0 {
		ttl = refreshTTL
	}
	if err := s.client.Set(ctx, key, marked, ttl).Err(); err != nil {
		return zero, zeroData, fmt.Errorf("redis refresh mark used: %w", err)
	}

	// Drop previous access session.
	if old.AccessToken != "" {
		_ = s.client.Del(ctx, sessionPrefix+old.AccessToken).Err()
	}

	if err := s.writePair(ctx, data, old.FamilyID, newAccess, newRefresh, accessTTL, refreshTTL); err != nil {
		return zero, zeroData, err
	}

	return ports.TokenPair{AccessToken: newAccess, RefreshToken: newRefresh}, data, nil
}

// Revoke deletes the access session.
func (s *RedisStore) Revoke(ctx context.Context, accessToken string) error {
	if accessToken == "" {
		return nil
	}

	raw, err := s.client.Get(ctx, sessionPrefix+accessToken).Bytes()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("redis session get: %w", err)
	}

	var sess redisSession
	if err := json.Unmarshal(raw, &sess); err != nil {
		return fmt.Errorf("redis session decode: %w", err)
	}

	if err := s.client.Del(ctx, sessionPrefix+accessToken).Err(); err != nil {
		return fmt.Errorf("redis session del: %w", err)
	}
	if sess.FamilyID != "" {
		return s.revokeFamily(ctx, sess.FamilyID)
	}
	return nil
}

// RevokeByRefresh revokes the refresh token's family.
func (s *RedisStore) RevokeByRefresh(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	raw, err := s.client.Get(ctx, refreshPrefix+refreshToken).Bytes()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("redis refresh get: %w", err)
	}

	var ref redisRefresh
	if err := json.Unmarshal(raw, &ref); err != nil {
		return fmt.Errorf("redis refresh decode: %w", err)
	}
	return s.revokeFamily(ctx, ref.FamilyID)
}

func (s *RedisStore) writePair(
	ctx context.Context,
	data ports.SessionData,
	familyID, access, refresh string,
	accessTTL, refreshTTL time.Duration,
) error {
	sessPayload, err := json.Marshal(redisSession{
		UserID:   data.UserID,
		Username: data.Username,
		Email:    data.Email,
		FamilyID: familyID,
	})
	if err != nil {
		return fmt.Errorf("redis session encode: %w", err)
	}

	refPayload, err := json.Marshal(redisRefresh{
		UserID:      data.UserID,
		Username:    data.Username,
		Email:       data.Email,
		FamilyID:    familyID,
		AccessToken: access,
	})
	if err != nil {
		return fmt.Errorf("redis refresh encode: %w", err)
	}

	pipe := s.client.TxPipeline()
	pipe.Set(ctx, sessionPrefix+access, sessPayload, accessTTL)
	pipe.Set(ctx, refreshPrefix+refresh, refPayload, refreshTTL)
	pipe.SAdd(ctx, familyPrefix+familyID+":tokens", refresh)
	pipe.Expire(ctx, familyPrefix+familyID+":tokens", refreshTTL)
	// Ensure family is marked active (clears prior revoked flag on new login).
	pipe.Del(ctx, familyPrefix+familyID+":revoked")
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis session write: %w", err)
	}
	return nil
}

func (s *RedisStore) familyRevoked(ctx context.Context, familyID string) (bool, error) {
	if familyID == "" {
		return false, nil
	}
	n, err := s.client.Exists(ctx, familyPrefix+familyID+":revoked").Result()
	if err != nil {
		return false, fmt.Errorf("redis family check: %w", err)
	}
	return n > 0, nil
}

func (s *RedisStore) revokeFamily(ctx context.Context, familyID string) error {
	if familyID == "" {
		return nil
	}

	tokensKey := familyPrefix + familyID + ":tokens"
	tokens, err := s.client.SMembers(ctx, tokensKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return fmt.Errorf("redis family members: %w", err)
	}

	pipe := s.client.TxPipeline()
	for _, t := range tokens {
		raw, err := s.client.Get(ctx, refreshPrefix+t).Bytes()
		if err == nil {
			var ref redisRefresh
			if json.Unmarshal(raw, &ref) == nil && ref.AccessToken != "" {
				pipe.Del(ctx, sessionPrefix+ref.AccessToken)
			}
		}
		pipe.Del(ctx, refreshPrefix+t)
	}
	pipe.Del(ctx, tokensKey)
	pipe.Set(ctx, familyPrefix+familyID+":revoked", "1", 7*24*time.Hour)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis family revoke: %w", err)
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
	_ auth.SessionLookup = (*RedisStore)(nil)
	_ ports.SessionStore = (*RedisStore)(nil)
)
