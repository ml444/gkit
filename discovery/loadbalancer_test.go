package discovery

import (
	"context"
	"errors"
	"testing"
)

func instances(ids ...string) []ServiceInstancer {
	out := make([]ServiceInstancer, len(ids))
	for i, id := range ids {
		out[i] = &ServiceInstance{ID: id, Name: "svc"}
	}
	return out
}

func TestLoadBalancers_Empty(t *testing.T) {
	ctx := context.Background()
	lbs := []LoadBalancer{
		NewRandomLoadBalancer(),
		NewRoundRobinLoadBalancer(),
		NewLeastConnectionsLoadBalancer(),
	}
	for _, lb := range lbs {
		_, err := lb.Select(ctx, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("expected ErrNotFound, got %v", err)
		}
	}
}

func TestRoundRobinLoadBalancer_Select(t *testing.T) {
	lb := NewRoundRobinLoadBalancer()
	ctx := context.Background()
	ins := instances("a", "b", "c")

	seen := make(map[string]int)
	for i := 0; i < 9; i++ {
		selected, err := lb.Select(ctx, ins)
		if err != nil {
			t.Fatalf("select: %v", err)
		}
		seen[selected.GetID()]++
	}
	for _, id := range []string{"a", "b", "c"} {
		if seen[id] != 3 {
			t.Fatalf("expected 3 selections for %s, got %d", id, seen[id])
		}
	}
}

func TestLeastConnectionsLoadBalancer_SelectUpdate(t *testing.T) {
	lb := NewLeastConnectionsLoadBalancer()
	ctx := context.Background()
	ins := instances("a", "b")

	first, err := lb.Select(ctx, ins)
	if err != nil {
		t.Fatalf("select first: %v", err)
	}
	second, err := lb.Select(ctx, ins)
	if err != nil {
		t.Fatalf("select second: %v", err)
	}
	if first.GetID() == second.GetID() {
		t.Fatalf("expected different instances for least connections")
	}

	lb.Update(ctx, first, true)
	third, err := lb.Select(ctx, ins)
	if err != nil {
		t.Fatalf("select third: %v", err)
	}
	if third.GetID() != first.GetID() {
		t.Fatalf("expected %s after update, got %s", first.GetID(), third.GetID())
	}
}

func TestLeastConnectionsLoadBalancer_PruneStale(t *testing.T) {
	lb := NewLeastConnectionsLoadBalancer()
	ctx := context.Background()

	_, _ = lb.Select(ctx, instances("a", "b"))
	_, err := lb.Select(ctx, instances("c"))
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	lbImpl := lb
	lbImpl.mu.Lock()
	defer lbImpl.mu.Unlock()
	if _, ok := lbImpl.connections["a"]; ok {
		t.Fatal("expected stale instance a to be pruned")
	}
	if _, ok := lbImpl.connections["c"]; !ok {
		t.Fatal("expected instance c to remain")
	}
}

func TestNewLoadBalancer(t *testing.T) {
	for _, typ := range []LoadBalancerType{RandomLoadBalancer, RoundRobinLoadBalancer, LeastConnectionsLoadBalancer} {
		lb, err := NewLoadBalancer(typ)
		if err != nil || lb == nil {
			t.Fatalf("NewLoadBalancer(%q) = %v, %v", typ, lb, err)
		}
	}
	if _, err := NewLoadBalancer(LoadBalancerType("invalid")); err == nil {
		t.Fatal("expected error for invalid load balancer type")
	}
}
