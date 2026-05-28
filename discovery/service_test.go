package discovery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestDefaultRegistry_RegisterDeregisterGet(t *testing.T) {
	reg := NewDefaultRegistry()
	ctx := context.Background()

	ins1 := &ServiceInstance{ID: "1", Name: "svc", Address: "127.0.0.1", Port: 8080}
	ins2 := &ServiceInstance{ID: "2", Name: "svc", Address: "127.0.0.2", Port: 8080}

	if err := reg.Register(ctx, nil); err == nil {
		t.Fatal("expected error for nil instance")
	}

	if err := reg.Register(ctx, ins1); err != nil {
		t.Fatalf("register ins1: %v", err)
	}
	if err := reg.Register(ctx, ins2); err != nil {
		t.Fatalf("register ins2: %v", err)
	}

	instances, err := reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	updated := &ServiceInstance{ID: "1", Name: "svc", Address: "10.0.0.1", Port: 9090}
	if err := reg.Register(ctx, updated); err != nil {
		t.Fatalf("register updated: %v", err)
	}
	instances, err = reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get after update: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances after update, got %d", len(instances))
	}
	if instances[0].GetAddress() != "10.0.0.1" && instances[1].GetAddress() != "10.0.0.1" {
		t.Fatalf("expected updated address")
	}

	if err := reg.Deregister(ctx, ins1); err != nil {
		t.Fatalf("deregister ins1: %v", err)
	}
	instances, err = reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get after deregister: %v", err)
	}
	if len(instances) != 1 || instances[0].GetID() != "2" {
		t.Fatalf("unexpected instances after deregister: %+v", instances)
	}

	if err := reg.Deregister(ctx, ins2); err != nil {
		t.Fatalf("deregister ins2: %v", err)
	}
	_, err = reg.GetServiceInstances(ctx, "svc")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDefaultRegistry_ConcurrentRegister(t *testing.T) {
	reg := NewDefaultRegistry()
	ctx := context.Background()

	const n = 100
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("id-%d", i)
			_ = reg.Register(ctx, &ServiceInstance{
				ID:      id,
				Name:    "svc",
				Address: "127.0.0.1",
				Port:    8080 + i,
			})
		}()
	}
	wg.Wait()

	instances, err := reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get instances: %v", err)
	}
	if len(instances) != n {
		t.Fatalf("expected %d instances, got %d", n, len(instances))
	}
}

func TestDefaultRegistry_Close(t *testing.T) {
	reg := NewDefaultRegistry()
	ctx := context.Background()
	_ = reg.Register(ctx, &ServiceInstance{ID: "1", Name: "svc", Address: "127.0.0.1", Port: 8080})
	if err := reg.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	_, err := reg.GetServiceInstances(ctx, "svc")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after close, got %v", err)
	}
}
