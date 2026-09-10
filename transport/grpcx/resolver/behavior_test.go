package resolver

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/ml444/gkit/discovery"
	gr "google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

type safeConn struct {
	mu      sync.Mutex
	states  []gr.State
	errs    []error
	updated chan struct{}
}

func (c *safeConn) UpdateState(s gr.State) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.states = append(c.states, s)
	if c.updated != nil {
		select {
		case c.updated <- struct{}{}:
		default:
		}
	}
	return nil
}
func (c *safeConn) ReportError(e error)                                { c.mu.Lock(); defer c.mu.Unlock(); c.errs = append(c.errs, e) }
func (*safeConn) NewAddress([]gr.Address)                              {}
func (*safeConn) ParseServiceConfig(string) *serviceconfig.ParseResult { return nil }
func (c *safeConn) counts() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.states), len(c.errs)
}

type registryFunc struct {
	discovery.ServiceRegistry
	get func(context.Context, string) ([]discovery.ServiceInstancer, error)
}

func (r registryFunc) GetServiceInstances(ctx context.Context, name string) ([]discovery.ServiceInstancer, error) {
	return r.get(ctx, name)
}

func buildTestResolver(t *testing.T, b *Builder, service string) (gr.Resolver, *safeConn) {
	t.Helper()
	cc := &safeConn{updated: make(chan struct{}, 1)}
	r, err := b.Build(gr.Target{URL: url.URL{Scheme: "discovery", Path: "/" + service}}, cc, gr.BuildOptions{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r.Close)
	select {
	case <-cc.updated:
	case <-time.After(time.Second):
		t.Fatal("initial resolution timed out")
	}
	return r, cc
}

func TestBuilderIsolatesServicesAndRegistries(t *testing.T) {
	ctx := context.Background()
	a := discovery.NewDefaultRegistry()
	b := discovery.NewDefaultRegistry()
	first := &discovery.ServiceInstance{ID: "a", Name: "svc", Address: "127.0.0.1", Port: 8001}
	other := &discovery.ServiceInstance{ID: "other", Name: "other", Address: "127.0.0.1", Port: 8003}
	second := &discovery.ServiceInstance{ID: "b", Name: "svc", Address: "127.0.0.1", Port: 8001} // same address, distinct registry identity
	for _, inst := range []*discovery.ServiceInstance{first, other} {
		if err := a.Register(ctx, inst); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.Register(ctx, second); err != nil {
		t.Fatal(err)
	}
	ba := NewBuilder(discovery.NewDiscoveryClient(a, discovery.WithCacheTTL(0)))
	bb := NewBuilder(discovery.NewDiscoveryClient(b, discovery.WithCacheTTL(0)))
	ra, _ := buildTestResolver(t, ba, "svc")
	ro, _ := buildTestResolver(t, ba, "other")
	buildTestResolver(t, bb, "svc")
	if inst, ok := ba.GetInstance("svc", "127.0.0.1:8001"); !ok || inst.GetID() != "a" {
		t.Fatalf("registry a instance: %v", inst)
	}
	if inst, ok := bb.GetInstance("svc", "127.0.0.1:8001"); !ok || inst.GetID() != "b" {
		t.Fatalf("registry b instance: %v", inst)
	}
	if _, ok := ba.GetInstance("other", "127.0.0.1:8001"); ok {
		t.Fatal("cross-service cache lookup")
	}
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			ra.ResolveNow(gr.ResolveNowOptions{})
			ba.GetInstance("svc", "127.0.0.1:8001")
		}()
		go func() {
			defer wg.Done()
			ro.ResolveNow(gr.ResolveNowOptions{})
			ba.GetInstance("other", "127.0.0.1:8003")
		}()
	}
	wg.Wait()
	if err := a.Deregister(ctx, first); err != nil {
		t.Fatal(err)
	}
	ra.ResolveNow(gr.ResolveNowOptions{})
	if _, ok := ba.GetInstance("svc", "127.0.0.1:8001"); ok {
		t.Fatal("offline instance retained")
	}
	if _, ok := ba.GetInstance("other", "127.0.0.1:8003"); !ok {
		t.Fatal("other service cache erased")
	}
	ro.Close()
	ro.Close()
	if _, ok := ba.GetInstance("other", "127.0.0.1:8003"); ok {
		t.Fatal("closed resolver retained cache")
	}
}

func TestResolverTransientErrorAndAddressChange(t *testing.T) {
	inst := &discovery.ServiceInstance{ID: "one", Name: "svc", Address: "::1", Port: 8001}
	var resultErr error
	reg := registryFunc{get: func(context.Context, string) ([]discovery.ServiceInstancer, error) {
		return []discovery.ServiceInstancer{inst}, resultErr
	}}
	b := NewBuilder(discovery.NewDiscoveryClient(reg, discovery.WithCacheTTL(0)))
	r, cc := buildTestResolver(t, b, "svc")
	resultErr = errors.New("temporary failure")
	r.ResolveNow(gr.ResolveNowOptions{})
	if _, ok := b.GetInstance("svc", "[::1]:8001"); !ok {
		t.Fatal("transient error erased known address")
	}
	if states, errs := cc.counts(); states != 1 || errs != 1 {
		t.Fatalf("states=%d errors=%d", states, errs)
	}
	resultErr = nil
	inst = &discovery.ServiceInstance{ID: "two", Name: "svc", Address: "::1", Port: 8002}
	r.ResolveNow(gr.ResolveNowOptions{})
	if _, ok := b.GetInstance("svc", "[::1]:8001"); ok {
		t.Fatal("old address retained")
	}
	if _, ok := b.GetInstance("svc", "[::1]:8002"); !ok {
		t.Fatal("new address absent")
	}
}

func TestResolverCloseCancelsRefresh(t *testing.T) {
	entered := make(chan struct{})
	var mu sync.Mutex
	calls := 0
	reg := registryFunc{get: func(ctx context.Context, _ string) ([]discovery.ServiceInstancer, error) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		if n == 1 {
			return []discovery.ServiceInstancer{&discovery.ServiceInstance{ID: "one", Name: "svc", Address: "127.0.0.1", Port: 80}}, nil
		}
		close(entered)
		<-ctx.Done()
		return nil, ctx.Err()
	}}
	b := NewBuilder(discovery.NewDiscoveryClient(reg, discovery.WithCacheTTL(0)))
	r, cc := buildTestResolver(t, b, "svc")
	refreshed := make(chan struct{})
	go func() { r.ResolveNow(gr.ResolveNowOptions{}); close(refreshed) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("refresh did not enter registry")
	}
	closed := make(chan struct{})
	go func() { r.Close(); close(closed) }()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close failed to cancel discovery")
	}
	<-refreshed
	r.ResolveNow(gr.ResolveNowOptions{})
	if states, errs := cc.counts(); states != 1 || errs != 0 {
		t.Fatalf("update after close: states=%d errs=%d", states, errs)
	}
	if _, ok := b.GetInstance("svc", "127.0.0.1:80"); ok {
		t.Fatal("closed resolver leaked cache")
	}
}
