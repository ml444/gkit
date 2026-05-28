package discovery

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

type fakeRegistry struct {
	instances []ServiceInstancer
	getCalls  int32
	getErr    error
}

func (f *fakeRegistry) Register(ctx context.Context, instance ServiceInstancer) error {
	return nil
}

func (f *fakeRegistry) Deregister(ctx context.Context, instance ServiceInstancer) error {
	return nil
}

func (f *fakeRegistry) GetServiceInstances(ctx context.Context, serviceName string) ([]ServiceInstancer, error) {
	atomic.AddInt32(&f.getCalls, 1)
	if f.getErr != nil {
		return nil, f.getErr
	}
	if len(f.instances) == 0 {
		return nil, ErrNotFound
	}
	return f.instances, nil
}

func (f *fakeRegistry) Close() error { return nil }

type countingLB struct {
	last []ServiceInstancer
}

func (c *countingLB) Select(ctx context.Context, instances []ServiceInstancer) (ServiceInstancer, error) {
	c.last = instances
	if len(instances) == 0 {
		return nil, ErrNotFound
	}
	return instances[0], nil
}

func (c *countingLB) Update(ctx context.Context, instance ServiceInstancer, success bool) {}

func testInstances(ids ...string) []ServiceInstancer {
	out := make([]ServiceInstancer, len(ids))
	for i, id := range ids {
		out[i] = &ServiceInstance{ID: id, Name: "svc"}
	}
	return out
}

func TestDiscoveryClient_CacheHitMiss(t *testing.T) {
	reg := &fakeRegistry{
		instances: testInstances("1"),
	}
	lb := &countingLB{}
	client := NewDiscoveryClient(reg,
		WithLoadBalancer(lb),
		WithCacheTTL(time.Minute),
	)
	ctx := context.Background()

	if _, err := client.GetServiceInstance(ctx, "svc"); err != nil {
		t.Fatalf("first get: %v", err)
	}
	if atomic.LoadInt32(&reg.getCalls) != 1 {
		t.Fatalf("expected 1 registry call, got %d", reg.getCalls)
	}

	if _, err := client.GetServiceInstance(ctx, "svc"); err != nil {
		t.Fatalf("second get: %v", err)
	}
	if atomic.LoadInt32(&reg.getCalls) != 1 {
		t.Fatalf("expected cache hit, registry calls=%d", reg.getCalls)
	}
}

func TestDiscoveryClient_CacheExpiry(t *testing.T) {
	reg := &fakeRegistry{
		instances: testInstances("1"),
	}
	client := NewDiscoveryClient(reg, WithCacheTTL(20*time.Millisecond))
	ctx := context.Background()

	if _, err := client.GetAllInstances(ctx, "svc"); err != nil {
		t.Fatalf("get: %v", err)
	}
	time.Sleep(30 * time.Millisecond)
	if _, err := client.GetAllInstances(ctx, "svc"); err != nil {
		t.Fatalf("get after expiry: %v", err)
	}
	if atomic.LoadInt32(&reg.getCalls) != 2 {
		t.Fatalf("expected 2 registry calls after expiry, got %d", reg.getCalls)
	}
}

func TestDiscoveryClient_GetAllInstancesReturnsCopy(t *testing.T) {
	reg := &fakeRegistry{
		instances: testInstances("1"),
	}
	client := NewDiscoveryClient(reg, WithCacheTTL(time.Minute))
	ctx := context.Background()

	first, err := client.GetAllInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	first[0] = &ServiceInstance{ID: "mutated"}

	second, err := client.GetAllInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get cached: %v", err)
	}
	if second[0].GetID() != "1" {
		t.Fatalf("cache mutated by caller, got id=%s", second[0].GetID())
	}
}

func TestDiscoveryClient_RefreshCacheClearsEmpty(t *testing.T) {
	reg := &fakeRegistry{
		instances: testInstances("1"),
	}
	client := NewDiscoveryClient(reg, WithCacheTTL(time.Minute))
	ctx := context.Background()

	if _, err := client.GetAllInstances(ctx, "svc"); err != nil {
		t.Fatalf("prime cache: %v", err)
	}

	reg.instances = nil
	reg.getErr = ErrNotFound
	if err := client.RefreshCache(ctx, "svc"); err != nil {
		t.Fatalf("refresh: %v", err)
	}
	_, err := client.GetAllInstances(ctx, "svc")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after refresh cleared cache, got %v", err)
	}
}

func TestDiscoveryClient_DisabledCache(t *testing.T) {
	reg := &fakeRegistry{
		instances: testInstances("1"),
	}
	client := NewDiscoveryClient(reg, WithCacheTTL(0))
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		if _, err := client.GetServiceInstance(ctx, "svc"); err != nil {
			t.Fatalf("get %d: %v", i, err)
		}
	}
	if atomic.LoadInt32(&reg.getCalls) != 2 {
		t.Fatalf("expected cache disabled, registry calls=%d", reg.getCalls)
	}
}

func TestDiscoveryClient_RegistryError(t *testing.T) {
	reg := &fakeRegistry{getErr: ErrNotFound}
	client := NewDiscoveryClient(reg)
	_, err := client.GetServiceInstance(context.Background(), "svc")
	if err == nil || !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected wrapped ErrNotFound, got %v", err)
	}
}

func TestDiscoveryClient_ClearCache(t *testing.T) {
	reg := &fakeRegistry{
		instances: testInstances("1"),
	}
	client := NewDiscoveryClient(reg, WithCacheTTL(time.Minute))
	ctx := context.Background()

	if _, err := client.GetAllInstances(ctx, "svc"); err != nil {
		t.Fatalf("prime cache: %v", err)
	}
	client.ClearCache()
	if _, err := client.GetAllInstances(ctx, "svc"); err != nil {
		t.Fatalf("get after clear: %v", err)
	}
	if atomic.LoadInt32(&reg.getCalls) != 2 {
		t.Fatalf("expected cache cleared, registry calls=%d", reg.getCalls)
	}
}
