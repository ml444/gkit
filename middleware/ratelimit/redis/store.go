// Package redis provides a Redis-backed ratelimit.Store for distributed rate limiting.
//
// Import this optional submodule when you need cluster-wide limits; the main gkit
// module does not require go-redis.
//
// Redis key format (see ratelimit.RateLimitKey):
//
//	gkit:rl:{service}:{path}:{windowMs}
//
// Example:
//
//	store := redis.NewStore(client, redis.Config{Service: "user-api"})
//	mw := ratelimit.FrequencyLimitWithStore(store, cfgs, ratelimit.WithServiceName("user-api"))
package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/ml444/gkit/middleware/ratelimit"
)

// Config configures the Redis rate limit store.
type Config struct {
	// Service names the limit namespace segment (defaults to "default").
	Service string
	// Prefix overrides the key prefix (default ratelimit.DefaultKeyPrefix).
	Prefix string
}

// Store implements ratelimit.Store with Redis fixed-window counters.
type Store struct {
	client *goredis.Client
	cfg    Config
}

// NewStore creates a Redis-backed store. Keys use ratelimit.RateLimitKey when
// FrequencyLimitWithStore passes the full key; Allow also accepts pre-built keys.
func NewStore(client *goredis.Client, cfg Config) *Store {
	if cfg.Service == "" {
		cfg.Service = "default"
	}
	return &Store{client: client, cfg: cfg}
}

// Allow increments the counter for key and returns whether the limit is exceeded.
// Prefer keys from ratelimit.RateLimitKey(service, path, period).
func (s *Store) Allow(ctx context.Context, key string, period time.Duration, limit uint64) (bool, error) {
	if period <= 0 || limit == 0 {
		return true, nil
	}
	pipe := s.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, period)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	count, err := incr.Result()
	if err != nil {
		return false, err
	}
	return uint64(count) <= limit, nil
}

// BuildKey is a helper matching gkit:rl:{service}:{path}:{windowMs}.
func (s *Store) BuildKey(path string, period time.Duration) string {
	return ratelimit.RateLimitKey(s.cfg.Service, path, period)
}
