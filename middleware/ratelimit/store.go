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
	mu      sync.Mutex
	windows map[string]*memWindow
	stopCh  chan struct{} // 用于优雅关闭清理协程
}

type memWindow struct {
	period int64
	limit  uint64
	start  int64
	count  uint64
}

// NewMemoryStore creates an in-memory rate limit store.
func NewMemoryStore(cleanupInterval time.Duration) *MemoryStore {
	s := &MemoryStore{windows: make(map[string]*memWindow)}
	go s.cleanup(cleanupInterval)
	return s
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

// cleanup 定期清理过期的限流记录
func (s *MemoryStore) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now().UnixMilli()
			s.mu.Lock()
			for k, w := range s.windows {
				// 如果当前时间已经超过了该窗口的起始时间 + 周期，说明该窗口已经过期
				if now-w.start >= w.period {
					delete(s.windows, k)
				}
			}
			s.mu.Unlock()
		case <-s.stopCh:
			return
		}
	}
}

// Close 关闭 Store 时停止清理协程
func (s *MemoryStore) Close() {
	close(s.stopCh)
}
