package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Store checks whether a key is allowed for the given period/limit window.
type Store interface {
	Allow(ctx context.Context, key string, period time.Duration, limit uint64) (bool, error)
}

// MemoryStore is an in-process fixed-window limiter store.
type MemoryStore struct {
	mu    sync.Mutex
	windows map[string]*memWindow
}

type memWindow struct {
	period    int64
	limit     uint64
	start     int64
	count     uint64
}

// NewMemoryStore creates an in-memory rate limit store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{windows: make(map[string]*memWindow)}
}

func (s *MemoryStore) Allow(_ context.Context, key string, period time.Duration, limit uint64) (bool, error) {
	now := time.Now().UnixMilli()
	periodMs := period.Milliseconds()
	if periodMs <= 0 || limit == 0 {
		return true, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.windows[key]
	if !ok || w.period != periodMs || w.limit != limit {
		w = &memWindow{period: periodMs, limit: limit, start: now, count: 1}
		s.windows[key] = w
		return true, nil
	}
	if now-w.start >= periodMs {
		w.start = now
		w.count = 1
		return true, nil
	}
	w.count++
	return w.count <= limit, nil
}
