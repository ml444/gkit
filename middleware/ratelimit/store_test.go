package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestMemoryStore_Allow(t *testing.T) {
	s := NewMemoryStore()
	ctx := context.Background()
	key := "test"
	period := 100 * time.Millisecond
	limit := uint64(2)
	for i := 0; i < 2; i++ {
		ok, err := s.Allow(ctx, key, period, limit)
		if err != nil || !ok {
			t.Fatalf("attempt %d: ok=%v err=%v", i, ok, err)
		}
	}
	ok, err := s.Allow(ctx, key, period, limit)
	if err != nil || ok {
		t.Fatalf("third attempt should be denied: ok=%v err=%v", ok, err)
	}
}
