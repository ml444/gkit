package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"github.com/ml444/gkit/discovery"
)

func setupRedisRegistry(t *testing.T, ttl int64) (*RedisRegistry, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	reg, err := NewRedisRegistryWithClient(client, WithTTL(ttl))
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	return reg, mr
}

func TestRedisRegistry_RegisterGetDeregister(t *testing.T) {
	reg, mr := setupRedisRegistry(t, 60)
	defer reg.Close()
	defer mr.Close()

	ctx := context.Background()
	ins := &discovery.ServiceInstance{
		ID: "1", Name: "svc", Address: "127.0.0.1", Port: 8080,
	}
	if err := reg.Register(ctx, ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	instances, err := reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 || instances[0].GetAddress() != "127.0.0.1" {
		t.Fatalf("unexpected instances: %+v", instances)
	}

	if err := reg.Deregister(ctx, ins); err != nil {
		t.Fatalf("deregister: %v", err)
	}
	_, err = reg.GetServiceInstances(ctx, "svc")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRedisRegistry_ExpireAndCleanupStaleMember(t *testing.T) {
	reg, mr := setupRedisRegistry(t, 1)
	defer reg.Close()
	defer mr.Close()

	ctx := context.Background()
	ins := &discovery.ServiceInstance{ID: "1", Name: "svc", Address: "127.0.0.1", Port: 8080}
	if err := reg.Register(ctx, ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	mr.FastForward(2 * time.Second)
	reg.serviceMap.Delete("svc")

	_, err := reg.GetServiceInstances(ctx, "svc")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after TTL, got %v", err)
	}
}

func TestRedisRegistry_CloseStopsBackgroundWorker(t *testing.T) {
	reg, mr := setupRedisRegistry(t, 60)
	mr.Close()
	if err := reg.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}
