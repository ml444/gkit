package grpcx

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/ml444/gkit/discovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	hp "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type feedbackRecord struct {
	mu      sync.Mutex
	ids     []string
	success []bool
}

func (*feedbackRecord) Select(_ context.Context, instances []discovery.ServiceInstancer) (discovery.ServiceInstancer, error) {
	return instances[0], nil
}
func (r *feedbackRecord) Update(_ context.Context, inst discovery.ServiceInstancer, success bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ids = append(r.ids, inst.GetID())
	r.success = append(r.success, success)
}

type outcomeHealth struct {
	hp.UnimplementedHealthServer
	value hp.HealthCheckResponse_ServingStatus
}

func (h outcomeHealth) Check(_ context.Context, req *hp.HealthCheckRequest) (*hp.HealthCheckResponse, error) {
	switch req.Service {
	case "unavailable":
		return nil, status.Error(codes.Unavailable, "backend unavailable")
	case "deadline":
		return nil, status.Error(codes.DeadlineExceeded, "backend deadline")
	case "business":
		return nil, status.Error(codes.InvalidArgument, "bad input")
	}
	return &hp.HealthCheckResponse{Status: h.value}, nil
}

func TestClientsWithSameServiceUseOwnRegistryAndFeedback(t *testing.T) {
	var checks []func()
	for _, tc := range []struct {
		id    string
		value hp.HealthCheckResponse_ServingStatus
	}{{"a", hp.HealthCheckResponse_SERVING}, {"b", hp.HealthCheckResponse_NOT_SERVING}} {
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		srv := grpc.NewServer()
		hp.RegisterHealthServer(srv, outcomeHealth{value: tc.value})
		go srv.Serve(lis)
		t.Cleanup(srv.Stop)
		host, portString, _ := net.SplitHostPort(lis.Addr().String())
		port, _ := strconv.Atoi(portString)
		reg := discovery.NewDefaultRegistry()
		inst := &discovery.ServiceInstance{ID: tc.id, Name: "shared-service", Address: host, Port: port}
		if err = reg.Register(context.Background(), inst); err != nil {
			t.Fatal(err)
		}
		feedback := new(feedbackRecord)
		dc := discovery.NewDiscoveryClient(reg, discovery.WithLoadBalancer(feedback))
		c, err := NewClient(WithEndpoint("discovery:///shared-service"), WithDiscovery(dc, ""))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Close() })
		// Run after both clients have been constructed, to catch global resolver reuse.
		checks = append(checks, func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			var p peer.Peer
			out, err := hp.NewHealthClient(c.Conn()).Check(ctx, &hp.HealthCheckRequest{}, grpc.Peer(&p))
			if err != nil || out.GetStatus() != tc.value {
				t.Fatalf("client %s: out=%v error=%v", tc.id, out, err)
			}
			if p.Addr == nil || p.Addr.String() != lis.Addr().String() {
				t.Fatalf("caller peer option not preserved: %v", p.Addr)
			}
			for _, entry := range []struct {
				service string
				code    codes.Code
			}{{"unavailable", codes.Unavailable}, {"deadline", codes.DeadlineExceeded}, {"business", codes.InvalidArgument}} {
				_, err = hp.NewHealthClient(c.Conn()).Check(ctx, &hp.HealthCheckRequest{Service: entry.service})
				if status.Code(err) != entry.code {
					t.Fatalf("original RPC error replaced: %v", err)
				}
			}
			feedback.mu.Lock()
			defer feedback.mu.Unlock()
			if len(feedback.ids) != 4 {
				t.Fatalf("feedback count=%v", feedback.ids)
			}
			for i, id := range feedback.ids {
				if id != tc.id {
					t.Errorf("cross-client feedback: %s", id)
				}
				want := i == 0 || i == 3
				if feedback.success[i] != want {
					t.Errorf("feedback %d=%v", i, feedback.success[i])
				}
			}
		})
	}
	for _, check := range checks {
		check()
	}
}

func TestFeedbackWithoutPeerPreservesOriginalError(t *testing.T) {
	c := &Client{discovery: discovery.NewDiscoveryClient(discovery.NewDefaultRegistry()), service: "svc"}
	cause := errors.New("dial failed")
	err := c.discoveryFeedbackInterceptor()(context.Background(), "/svc/Check", nil, nil, nil, func(context.Context, string, interface{}, interface{}, *grpc.ClientConn, ...grpc.CallOption) error {
		return cause
	})
	if err != cause {
		t.Fatalf("original error replaced: %v", err)
	}
}

func TestTimeoutOptionsAndCallerDeadline(t *testing.T) {
	c := &Client{}
	WithTimeout(time.Second)(c)
	WithDialTimeout(0)(c)
	WithCallTimeout(2 * time.Second)(c)
	if c.timeout != 2*time.Second || c.dialTimeout != 0 {
		t.Fatalf("timeouts=%v %v", c.timeout, c.dialTimeout)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	want, _ := ctx.Deadline()
	_, err := NewClient(WithEndpoint("buf"), WithCallTimeout(-1))
	if err == nil {
		t.Fatal("negative call timeout accepted")
	}
	_, err = NewClient(WithEndpoint("buf"), WithDialTimeout(-1))
	if err == nil {
		t.Fatal("negative dial timeout accepted")
	}
	err = c.timeoutInterceptor()(ctx, "/svc/Check", nil, nil, nil, func(ctx context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		got, _ := ctx.Deadline()
		if !got.Equal(want) {
			t.Errorf("caller deadline changed: %v want %v", got, want)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	WithCallTimeout(0)(c)
	err = c.timeoutInterceptor()(context.Background(), "/svc/Check", nil, nil, nil, func(ctx context.Context, _ string, _, _ interface{}, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		if _, ok := ctx.Deadline(); ok {
			t.Error("disabled timeout still sets deadline")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestZeroDialTimeoutAndStreaming(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	srv := grpc.NewServer()
	hs := health.NewServer()
	hp.RegisterHealthServer(srv, hs)
	go srv.Serve(lis)
	defer srv.Stop()
	c, err := NewClient(WithEndpoint("buf"), WithTimeout(0), WithDialOptions(grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) })))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err = hp.NewHealthClient(c.Conn()).Check(context.Background(), &hp.HealthCheckRequest{}); err != nil {
		t.Fatal(err)
	}
	streamClient, err := NewClient(WithEndpoint("buf"), WithCallTimeout(time.Millisecond), WithDialOptions(grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) })))
	if err != nil {
		t.Fatal(err)
	}
	defer streamClient.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := hp.NewHealthClient(streamClient.Conn()).Watch(ctx, &hp.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = stream.Recv(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	hs.SetServingStatus("", hp.HealthCheckResponse_NOT_SERVING)
	if out, err := stream.Recv(); err != nil || out.GetStatus() != hp.HealthCheckResponse_NOT_SERVING {
		t.Fatalf("unary timeout affected stream: %v %v", out, err)
	}
	_, err = NewClient(WithEndpoint("blocked"), WithDialTimeout(20*time.Millisecond), WithDialOptions(grpc.WithBlock(), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { <-ctx.Done(); return nil, ctx.Err() })))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("blocking dial timeout: %v", err)
	}
}

func TestStopBeforeServingReleasesListener(t *testing.T) {
	for _, setup := range []func() (*Server, error){func() (*Server, error) { return NewServer() }, func() (*Server, error) {
		s, err := NewServer(Network("invalid"))
		if err == nil {
			_ = s.Start()
		}
		return s, err
	}} {
		s, err := setup()
		if err != nil {
			t.Fatal(err)
		}
		if err = s.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
		if err = s.Stop(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	s, err := NewServer(Address("127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}
	endpoint, err := s.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	lis, err := net.Listen("tcp", endpoint)
	if err != nil {
		t.Fatalf("early listener not released: %v", err)
	}
	lis.Close()
}

type blockingDiscoveryRegistry struct{ discovery.ServiceRegistry }

func (blockingDiscoveryRegistry) GetServiceInstances(ctx context.Context, _ string) ([]discovery.ServiceInstancer, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
func TestDialTimeoutCancelsInitialDiscovery(t *testing.T) {
	dc := discovery.NewDiscoveryClient(blockingDiscoveryRegistry{})
	start := time.Now()
	_, err := NewClient(WithEndpoint("discovery:///slow"), WithDiscovery(dc, ""), WithDialTimeout(30*time.Millisecond), WithDialOptions(grpc.WithBlock()))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("dial error=%v", err)
	}
	if time.Since(start) > time.Second {
		t.Fatal("initial discovery bypassed dial timeout")
	}
}
