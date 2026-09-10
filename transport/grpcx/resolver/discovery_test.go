package resolver

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"

	"github.com/ml444/gkit/discovery"
)

type recordingConn struct {
	states []resolver.State
	mu     sync.Mutex
}

func (r *recordingConn) UpdateState(s resolver.State) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.states = append(r.states, s)
	return nil
}

func (r *recordingConn) ReportError(error) {}
func (r *recordingConn) NewAddress([]resolver.Address) {
	// deprecated
}
func (r *recordingConn) ParseServiceConfig(string) *serviceconfig.ParseResult {
	return &serviceconfig.ParseResult{}
}

func TestDiscoveryResolver_UpdateState(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	reg := discovery.NewDefaultRegistry()
	dc := discovery.NewDiscoveryClient(reg, discovery.WithCacheTTL(time.Minute))
	if err := reg.Register(context.Background(), &discovery.ServiceInstance{
		ID: "i1", Name: "svc", Address: "127.0.0.1", Port: port,
	}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(context.Background(), &discovery.ServiceInstance{
		ID: "i2", Name: "svc", Address: "127.0.0.1", Port: port + 1,
	}); err != nil {
		t.Fatal(err)
	}

	Register(dc)
	cc := &recordingConn{}
	target := resolver.Target{
		URL: url.URL{Scheme: scheme, Path: "/svc"},
	}
	r, err := (&discoveryBuilder{dc: dc}).Build(target, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	defer r.Close()

	deadline := time.After(time.Second)
	for {
		cc.mu.Lock()
		ready := len(cc.states) > 0
		cc.mu.Unlock()
		if ready {
			break
		}
		select {
		case <-deadline:
			t.Fatal("initial resolution timed out")
		case <-time.After(time.Millisecond):
		}
	}
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if len(cc.states) == 0 {
		t.Fatal("expected UpdateState to be called")
	}
	last := cc.states[len(cc.states)-1]
	if got := len(last.Addresses); got != 2 {
		t.Fatalf("expected 2 addresses, got %d", got)
	}
}

func TestParseServiceName(t *testing.T) {
	target := resolver.Target{URL: url.URL{Scheme: scheme, Path: "/userService"}}
	if got := parseServiceName(target); got != "userService" {
		t.Fatalf("got %q", got)
	}
	target = resolver.Target{URL: url.URL{Scheme: scheme}}
	if got := parseServiceName(target); got != "" {
		t.Fatalf("empty target got %q", got)
	}
}

func TestInstanceFromAttributes(t *testing.T) {
	if got := InstanceFromAttributes(nil); got != nil {
		t.Fatalf("nil attrs = %#v", got)
	}
	if got := InstanceFromAttributes(attributes.New("other", "x")); got != nil {
		t.Fatalf("missing attrs = %#v", got)
	}
	inst := &discovery.ServiceInstance{ID: "i1", Name: "svc", Address: "127.0.0.1", Port: 80}
	if got := InstanceFromAttributes(attributes.New(instanceAttrKey, inst)); got != inst {
		t.Fatalf("instance = %#v", got)
	}
}

func TestDiscoveryBuilderErrorsAndResolveNow(t *testing.T) {
	reg := discovery.NewDefaultRegistry()
	dc := discovery.NewDiscoveryClient(reg, discovery.WithCacheTTL(time.Millisecond))
	// reg.Register(context.TODO(), &discovery.ServiceInstance{
	// 	ID:          "1",
	// 	Name:        "missing",
	// 	Version:     "1",
	// 	Address:     "",
	// 	Port:        0,
	// 	Metadata:    map[string]string{},
	// 	HealthCheck: "",
	// })
	Register(dc)
	if _, err := (&discoveryBuilder{dc: dc}).Build(resolver.Target{URL: url.URL{Scheme: scheme}}, &recordingConn{}, resolver.BuildOptions{}); err == nil {
		t.Fatal("expected empty service error")
	}

	cc := &recordingConn{}
	r := &discoveryResolver{
		dc:      dc,
		service: "missing",
		cc:      cc,
		refresh: time.Hour,
		stop:    make(chan struct{}),
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	if len(cc.states) != 1 || len(cc.states[0].Addresses) != 0 {
		t.Fatalf("states = %#v", cc.states)
	}
}
