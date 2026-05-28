//go:build integration

package redis

import (
	"context"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/ml444/gkit/discovery"
)

func TestRedisRegistry_IntegrationConcurrentGet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	reg, err := NewRedisRegistryWithClient(client)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		ins := &discovery.ServiceInstance{
			ID:      string(rune('a' + i)),
			Name:    "svc",
			Address: "127.0.0.1",
			Port:    8080 + i,
		}
		if err := reg.Register(ctx, ins); err != nil {
			t.Fatalf("register: %v", err)
		}
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			instances, err := reg.GetServiceInstances(ctx, "svc")
			if err != nil || len(instances) != 3 {
				t.Errorf("get instances: err=%v len=%d", err, len(instances))
			}
		}()
	}
	wg.Wait()
}
