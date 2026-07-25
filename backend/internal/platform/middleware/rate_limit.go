package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ismailtemuroglu/discord/internal/platform/httpx"
)

// KeyFunc extracts the rate-limit bucket key from a request (IP or user id).
type KeyFunc func(r *http.Request) string

// TokenBucketLimiter is an in-memory token-bucket rate limiter keyed by caller.
type TokenBucketLimiter struct {
	rate     float64 // tokens replenished per second
	capacity float64
	buckets  sync.Map // map[string]*tokenBucket
}

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	lastRefill time.Time
	rate       float64
	capacity   float64
}

// NewTokenBucketLimiter creates a limiter that allows `rps` sustained requests
// per second with a burst of `burst` tokens.
func NewTokenBucketLimiter(rps float64, burst int) *TokenBucketLimiter {
	if rps <= 0 {
		rps = 20
	}
	if burst <= 0 {
		burst = int(rps)
	}
	return &TokenBucketLimiter{
		rate:     rps,
		capacity: float64(burst),
	}
}

// Allow consumes one token for key. Returns false and a Retry-After duration
// when the bucket is empty.
func (l *TokenBucketLimiter) Allow(key string) (bool, time.Duration) {
	raw, _ := l.buckets.LoadOrStore(key, &tokenBucket{
		tokens:     l.capacity,
		lastRefill: time.Now(),
		rate:       l.rate,
		capacity:   l.capacity,
	})
	b := raw.(*tokenBucket)

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	if b.tokens < 1 {
		needed := 1 - b.tokens
		retry := time.Duration(needed / b.rate * float64(time.Second))
		if retry < time.Second {
			retry = time.Second
		}
		return false, retry
	}

	b.tokens--
	return true, 0
}

// ClientIP returns the caller IP from X-Forwarded-For (nginx) or RemoteAddr.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimit returns middleware that rejects excess requests with 429 Too Many Requests.
// Default key is ClientIP; pass a custom KeyFunc to key by authenticated user id.
func RateLimit(limiter *TokenBucketLimiter, keyFn ...KeyFunc) func(http.Handler) http.Handler {
	extract := ClientIP
	if len(keyFn) > 0 && keyFn[0] != nil {
		extract = keyFn[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extract(r)
			if key == "" {
				key = "unknown"
			}

			allowed, retryAfter := limiter.Allow(key)
			if !allowed {
				secs := int(retryAfter.Seconds())
				if secs < 1 {
					secs = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(secs))
				_ = httpx.WriteError(w, httpx.ErrTooManyReqs)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
