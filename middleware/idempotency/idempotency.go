package idempotency

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

var ErrDuplicate = errorx.CreateError(409, 40901, "IDEMPOTENCY: duplicate request")

const headerKey = "Idempotency-Key"

// Store records idempotency keys.
type Store interface {
	Reserve(ctx context.Context, key string, ttl time.Duration) (bool, error)
}

// MemoryStore is an in-process idempotency store.
type MemoryStore struct {
	mu   sync.Mutex
	keys map[string]time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{keys: make(map[string]time.Time)}
}

func (s *MemoryStore) Reserve(_ context.Context, key string, ttl time.Duration) (bool, error) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, exp := range s.keys {
		if now.After(exp) {
			delete(s.keys, k)
		}
	}
	if exp, ok := s.keys[key]; ok && now.Before(exp) {
		return false, nil
	}
	s.keys[key] = now.Add(ttl)
	return true, nil
}

// HTTPMiddleware rejects duplicate Idempotency-Key within TTL.
func HTTPMiddleware(store Store, ttl time.Duration) middleware.HttpMiddleware {
	if store == nil {
		store = NewMemoryStore()
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get(headerKey)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			ok, err := store.Reserve(r.Context(), key, ttl)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, ErrDuplicate.Error(), http.StatusConflict)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Server is service-layer idempotency using context metadata key.
func Server(store Store, ttl time.Duration, keyFromContext func(context.Context) string) middleware.Middleware {
	if store == nil {
		store = NewMemoryStore()
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			key := ""
			if keyFromContext != nil {
				key = keyFromContext(ctx)
			}
			if key == "" {
				return next(ctx, req)
			}
			ok, err := store.Reserve(ctx, key, ttl)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, ErrDuplicate
			}
			return next(ctx, req)
		}
	}
}
