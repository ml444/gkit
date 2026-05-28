package resolver

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"

	"github.com/ml444/gkit/discovery"
)

type recordingConn struct {
	states []resolver.State
}

func (r *recordingConn) UpdateState(s resolver.State) error {
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
	r, err := discoveryBuilder{}.Build(target, cc, resolver.BuildOptions{})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	defer r.Close()

	time.Sleep(50 * time.Millisecond)
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
}
