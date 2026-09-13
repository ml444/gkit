package grpcx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ml444/gkit/transport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func lifecycleResult[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("lifecycle operation timed out")
		var zero T
		return zero
	}
}

func TestManagedOwnership(t *testing.T) {
	if _, err := Managed(nil, transport.LifecycleOptions{}); err == nil {
		t.Fatal("accepted nil server")
	}
	if _, err := Managed(&Server{}, transport.LifecycleOptions{}); err == nil {
		t.Fatal("accepted uninitialized server")
	}
	s, err := NewServer(Address("127.0.0.1:0"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Managed(s, transport.LifecycleOptions{ShutdownTimeout: -1}); err == nil {
		t.Fatal("accepted negative timeout")
	}
	ep, err := s.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	m, err := Managed(s, transport.LifecycleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Stop(context.Background())
	if _, err := Managed(s, transport.LifecycleOptions{}); !errors.Is(err, transport.ErrLifecycleOwned) {
		t.Fatalf("duplicate Managed: %v", err)
	}
	if _, err := s.Endpoint(); !errors.Is(err, transport.ErrLifecycleOwned) {
		t.Fatalf("legacy Endpoint after Managed: %v", err)
	}
	if err := s.Start(); !errors.Is(err, transport.ErrLifecycleOwned) {
		t.Fatalf("legacy Start after Managed: %v", err)
	}
	if err := s.Stop(context.Background()); !errors.Is(err, transport.ErrLifecycleOwned) {
		t.Fatalf("legacy Stop after Managed: %v", err)
	}
	if err := m.Listen(); err != nil {
		t.Fatal(err)
	}
	u, ok := m.EndpointURL()
	if !ok || u.Host != ep || u.Scheme != "grpc" {
		t.Fatalf("adopted endpoint: %v, %v", u, ok)
	}
	s, err = NewServer()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := Managed(s, transport.LifecycleOptions{}); !errors.Is(err, transport.ErrLifecycleOwned) {
		t.Fatalf("Managed after legacy Stop: %v", err)
	}
}

type closingListener struct {
	net.Listener
	closing chan struct{}
	once    sync.Once
}

func (l *closingListener) Close() error {
	l.once.Do(func() { close(l.closing) })
	return l.Listener.Close()
}

type lifecycleHealth struct {
	healthpb.UnimplementedHealthServer
	entered chan context.Context
	release chan struct{}
	done    chan struct{}
}

func (h *lifecycleHealth) wait(ctx context.Context) error {
	defer close(h.done)
	h.entered <- ctx
	select {
	case <-h.release:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (h *lifecycleHealth) Check(ctx context.Context, _ *healthpb.HealthCheckRequest) (*healthpb.HealthCheckResponse, error) {
	if err := h.wait(ctx); err != nil {
		return nil, err
	}
	return &healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING}, nil
}

func (h *lifecycleHealth) Watch(_ *healthpb.HealthCheckRequest, stream healthpb.Health_WatchServer) error {
	if err := h.wait(stream.Context()); err != nil {
		return err
	}
	return stream.Send(&healthpb.HealthCheckResponse{Status: healthpb.HealthCheckResponse_SERVING})
}

func TestManagedRPCGraceAndForcedStop(t *testing.T) {
	for _, streaming := range []bool{false, true} {
		for _, force := range []bool{false, true} {
			name, timeout := "unary/graceful", 2*time.Second
			if streaming {
				name = "stream/graceful"
			}
			if force {
				name += "/forced"
				timeout = 80 * time.Millisecond
			}
			t.Run(name, func(t *testing.T) {
				raw, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					t.Fatal(err)
				}
				defer raw.Close()
				lis := &closingListener{Listener: raw, closing: make(chan struct{})}
				s, err := NewServer(Address("127.0.0.1:0"), Listener(lis))
				if err != nil {
					t.Fatal(err)
				}
				h := &lifecycleHealth{entered: make(chan context.Context, 1), release: make(chan struct{}), done: make(chan struct{})}
				healthpb.RegisterHealthServer(s, h)
				m, err := Managed(s, transport.LifecycleOptions{ShutdownTimeout: timeout})
				if err != nil {
					t.Fatal(err)
				}
				defer m.Stop(context.Background())
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				start := make(chan error, 1)
				go func() { start <- m.Start(ctx) }()
				conn, err := grpc.Dial(raw.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
				rpcCtx, rpcCancel := context.WithTimeout(context.Background(), 4*time.Second)
				defer rpcCancel()
				result := make(chan error, 1)
				go func() {
					client := healthpb.NewHealthClient(conn)
					if streaming {
						stream, err := client.Watch(rpcCtx, &healthpb.HealthCheckRequest{})
						if err == nil {
							_, err = stream.Recv()
						}
						result <- err
					} else {
						_, err := client.Check(rpcCtx, &healthpb.HealthCheckRequest{})
						result <- err
					}
				}()
				requestCtx := lifecycleResult(t, h.entered)
				cancel()
				lifecycleResult(t, lis.closing)
				if !force {
					if err := requestCtx.Err(); err != nil {
						t.Fatalf("startup cancellation aborted active RPC: %v", err)
					}
					select {
					case err := <-start:
						t.Fatalf("Start returned with active RPC: %v", err)
					default:
					}
					close(h.release)
					if err := lifecycleResult(t, result); err != nil {
						t.Fatalf("RPC: %v", err)
					}
					if err := lifecycleResult(t, start); err != nil {
						t.Fatalf("graceful Start: %v", err)
					}
				} else {
					if err := lifecycleResult(t, start); !errors.Is(err, context.DeadlineExceeded) {
						t.Fatalf("forced Start: %v", err)
					}
					lifecycleResult(t, requestCtx.Done())
					if err := lifecycleResult(t, result); err == nil {
						t.Fatal("forced RPC unexpectedly completed")
					}
				}
				lifecycleResult(t, h.done)
			})
		}
	}
}

func TestManagedGRPCTLSAndHealthShutdown(t *testing.T) {
	for name, opts := range map[string][]ServerOption{
		"grpc":  {EnableHealth()},
		"tls":   {TLSConfig(&tls.Config{})},
		"creds": {Credentials(credentials.NewTLS(&tls.Config{}))},
	} {
		t.Run(name, func(t *testing.T) {
			s, err := NewServer(append(opts, Address("127.0.0.1:0"))...)
			if err != nil {
				t.Fatal(err)
			}
			m, err := Managed(s, transport.LifecycleOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer m.Stop(context.Background())
			if err := m.Listen(); err != nil {
				t.Fatal(err)
			}
			want := "grpcs"
			if name == "grpc" {
				want = "grpc"
			}
			u, ok := m.EndpointURL()
			if !ok || u.Scheme != want {
				t.Fatalf("endpoint: %v, %v", u, ok)
			}
			if err := m.Stop(context.Background()); err != nil {
				t.Fatal(err)
			}
			if s.health != nil {
				res, err := s.health.Check(context.Background(), &healthpb.HealthCheckRequest{Service: s.name})
				if err != nil || res.GetStatus() != healthpb.HealthCheckResponse_NOT_SERVING {
					t.Fatalf("health after Stop: %v, %v", res, err)
				}
			}
		})
	}
}
