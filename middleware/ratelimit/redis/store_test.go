package redis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"

	"github.com/ml444/gkit/middleware/ratelimit"
)

func TestStoreAllow(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := goredis.NewClient(&goredis.Options{Addr: mr.Addr()})
	store := NewStore(client, Config{Service: "api"})
	ctx := context.Background()
	key := ratelimit.RateLimitKey("api", "users/list", time.Second)
	for i := 0; i < 2; i++ {
		ok, err := store.Allow(ctx, key, time.Second, 2)
		if err != nil || !ok {
			t.Fatalf("attempt %d: ok=%v err=%v", i, ok, err)
		}
	}
	ok, err := store.Allow(ctx, key, time.Second, 2)
	if err != nil || ok {
		t.Fatalf("third should deny: ok=%v err=%v", ok, err)
	}
	if !mr.Exists(key) {
		t.Fatalf("redis key missing: %s", key)
	}
}
