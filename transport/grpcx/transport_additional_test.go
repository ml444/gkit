package grpcx

import (
	"context"
	"net"
	"testing"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func TestTransportAccessors(t *testing.T) {
	tr := &Transport{
		endpoint:  "127.0.0.1:1",
		operation: "/svc.Method",
		inMD:      transport.Pairs("in", "1"),
		outMD:     transport.Pairs("out", "2"),
	}
	if tr.Kind() != "grpc" || tr.Endpoint() != "127.0.0.1:1" || tr.Path() != "/svc.Method" {
		t.Fatalf("unexpected transport: %#v", tr)
	}
	if tr.In().GetFirst("in") != "1" || tr.Out().GetFirst("out") != "2" {
		t.Fatalf("metadata mismatch")
	}
	ctx := transport.ToContext(context.Background(), tr)
	if got, ok := GetTransport(ctx); !ok || got != tr {
		t.Fatalf("GetTransport = %#v %v", got, ok)
	}
	if got, ok := GetTransport(context.Background()); ok || got != nil {
		t.Fatalf("empty GetTransport = %#v %v", got, ok)
	}
}

func TestClientTransport(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()
	cc, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer cc.Close()

	tr := ClientTransport(context.Background(), "/svc/Method", cc, transport.Pairs("k", "v"))
	if tr.Endpoint() == "" || tr.Path() != "/svc/Method" || tr.In().GetFirst("k") != "v" || tr.Out() != nil {
		t.Fatalf("unexpected client transport: %#v", tr)
	}
}

func TestParseClientTarget(t *testing.T) {
	tests := []struct {
		endpoint string
		service  string
		target   string
		svc      string
		wantErr  bool
	}{
		{"", "", "", "", true},
		{"127.0.0.1:9000", "", "passthrough:///127.0.0.1:9000", "", false},
		{"discovery:///svc", "", "discovery:///svc", "svc", false},
		{"discovery:///svc", "override", "discovery:///override", "override", false},
		{"discovery:///", "", "", "", true},
	}
	for _, tt := range tests {
		target, svc, err := parseClientTarget(tt.endpoint, tt.service)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("parseClientTarget(%q) expected error", tt.endpoint)
			}
			continue
		}
		if err != nil || target != tt.target || svc != tt.svc {
			t.Fatalf("parseClientTarget(%q,%q) = %q %q %v", tt.endpoint, tt.service, target, svc, err)
		}
	}
}

func TestClientCloseAndPeerInstance(t *testing.T) {
	c := &Client{}
	if err := c.Close(); err != nil {
		t.Fatalf("close nil: %v", err)
	}
	reg := discovery.NewDefaultRegistry()
	dc := discovery.NewDiscoveryClient(reg)
	inst := &discovery.ServiceInstance{ID: "i1", Name: "svc", Address: "127.0.0.1", Port: 8080}
	if err := reg.Register(context.Background(), inst); err != nil {
		t.Fatal(err)
	}
	ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: fakeAddr("127.0.0.1:8080")})
	if got := instanceFromPeer(ctx, dc, "svc"); got == nil || got.GetID() != "i1" {
		t.Fatalf("instance = %#v", got)
	}
	if got := instanceFromPeer(context.Background(), dc, "svc"); got != nil {
		t.Fatalf("empty peer = %#v", got)
	}
	if got := instanceFromPeer(peer.NewContext(context.Background(), &peer.Peer{Addr: fakeAddr("bad")}), dc, "svc"); got != nil {
		t.Fatalf("bad peer = %#v", got)
	}
}

func TestDiscoveryFeedbackInterceptorBranches(t *testing.T) {
	c := &Client{}
	err := c.discoveryFeedbackInterceptor()(context.Background(), "/svc/Method", nil, nil, nil,
		func(context.Context, string, interface{}, interface{}, *grpc.ClientConn, ...grpc.CallOption) error {
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}

	reg := discovery.NewDefaultRegistry()
	dc := discovery.NewDiscoveryClient(reg)
	inst := &discovery.ServiceInstance{ID: "i1", Name: "svc", Address: "127.0.0.1", Port: 8081}
	if err = reg.Register(context.Background(), inst); err != nil {
		t.Fatal(err)
	}
	ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: fakeAddr("127.0.0.1:8081")})
	c = &Client{discovery: dc, service: "svc"}
	err = c.discoveryFeedbackInterceptor()(ctx, "/svc/Method", nil, nil, nil,
		func(context.Context, string, interface{}, interface{}, *grpc.ClientConn, ...grpc.CallOption) error {
			return nil
		})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewXDSConnError(t *testing.T) {
	t.Setenv("GRPC_XDS_BOOTSTRAP", "/path/that/does/not/exist")
	if _, err := NewXDSConn("xds:///listener"); err == nil {
		t.Fatal("expected xds error")
	}
}

type fakeAddr string

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return string(a) }
