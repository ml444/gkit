package transport_test

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/grpcx"
	"github.com/ml444/gkit/transport/httpx"
)

type managedFactory func(t *testing.T, network, address string, lis net.Listener) transport.ManagedServer

var managedFactories = map[string]managedFactory{
	"http": func(t *testing.T, network, address string, lis net.Listener) transport.ManagedServer {
		t.Helper()
		s := httpx.NewServer(httpx.Network(network), httpx.Address(address), httpx.Listener(lis))
		m, err := httpx.Managed(s, transport.LifecycleOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return m
	},
	"grpc": func(t *testing.T, network, address string, lis net.Listener) transport.ManagedServer {
		t.Helper()
		s, err := grpcx.NewServer(grpcx.Network(network), grpcx.Address(address), grpcx.Listener(lis))
		if err != nil {
			t.Fatal(err)
		}
		m, err := grpcx.Managed(s, transport.LifecycleOptions{})
		if err != nil {
			t.Fatal(err)
		}
		return m
	},
}

func managedStop(t *testing.T, m transport.ManagedServer) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := m.Stop(ctx); err != nil {
		t.Errorf("Stop: %v", err)
	}
}

func testListener(t *testing.T, network, address string) net.Listener {
	t.Helper()
	lis, err := net.Listen(network, address)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = lis.Close() })
	return lis
}

func TestManagedBindingContract(t *testing.T) {
	for name, factory := range managedFactories {
		t.Run(name, func(t *testing.T) {
			reserved := testListener(t, "tcp", "127.0.0.1:0")
			address := reserved.Addr().String()
			m := factory(t, "tcp", address, nil)
			defer managedStop(t, m)
			if u, ok := m.EndpointURL(); u != nil || ok {
				t.Fatalf("unbound endpoint: %v, %v", u, ok)
			}
			// Occupied port: query succeeds, explicit Listen fails and is retryable.
			if err := m.Listen(); err == nil {
				t.Fatal("Listen unexpectedly succeeded on occupied port")
			}
			_ = reserved.Close()
			m.EndpointURL()
			probe := testListener(t, "tcp", address)
			_ = probe.Close()
			var wg sync.WaitGroup
			for i := 0; i < 20; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if err := m.Listen(); err != nil {
						t.Errorf("Listen: %v", err)
					}
				}()
			}
			wg.Wait()
			u, ok := m.EndpointURL()
			if !ok || u.Host != address || u.Scheme != name {
				t.Fatalf("bound endpoint: %v, %v", u, ok)
			}
			u.Host = "modified"
			v, _ := m.EndpointURL()
			if v.Host != address {
				t.Fatal("URL snapshot changed server address")
			}
			managedStop(t, m)
			if u, ok := m.EndpointURL(); u == nil || ok {
				t.Fatalf("stopped endpoint: %v, %v", u, ok)
			}
			_ = testListener(t, "tcp", address).Close()
			if err := m.Listen(); !errors.Is(err, transport.ErrServerStopped) {
				t.Fatalf("Listen after Stop: %v", err)
			}
			if err := m.Start(context.Background()); !errors.Is(err, transport.ErrServerStopped) {
				t.Fatalf("Start after Stop: %v", err)
			}
		})
	}
}

func TestManagedStopBeforeServing(t *testing.T) {
	for name, factory := range managedFactories {
		for _, bind := range []bool{false, true} {
			label := "new"
			if bind {
				label = "bound"
			}
			t.Run(name+"/"+label, func(t *testing.T) {
				lis := testListener(t, "tcp", "127.0.0.1:0")
				m := factory(t, "tcp", "127.0.0.1:0", lis)
				if bind {
					if err := m.Listen(); err != nil {
						t.Fatal(err)
					}
					u, ok := m.EndpointURL()
					if !ok || u.Host != lis.Addr().String() {
						t.Fatalf("injected endpoint: %v, %v", u, ok)
					}
				}
				managedStop(t, m)
				managedStop(t, m)
				_ = testListener(t, "tcp", lis.Addr().String()).Close()
			})
		}
		t.Run(name+"/canceled-start", func(t *testing.T) {
			m := factory(t, "tcp", "127.0.0.1:0", nil)
			defer managedStop(t, m)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := m.Start(ctx); !errors.Is(err, context.Canceled) {
				t.Fatalf("initially canceled Start: %v", err)
			}
			if u, ok := m.EndpointURL(); u != nil || ok {
				t.Fatalf("canceled start allocated endpoint: %v, %v", u, ok)
			}
			if err := m.Listen(); err != nil {
				t.Fatal(err)
			}
			u, ok := m.EndpointURL()
			if !ok || u.Port() == "0" || u.Port() == "" {
				t.Fatalf("random port not resolved: %v", u)
			}
		})
	}
}

func TestManagedEndpointFailureClosesNewListener(t *testing.T) {
	for name, factory := range managedFactories {
		t.Run(name, func(t *testing.T) {
			// Endpoint extraction requires a TCP port. A Unix bind succeeds first,
			// then extraction fails: the newly allocated socket must be closed.
			dir, err := os.MkdirTemp("", "gkit-lc-")
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.RemoveAll(dir) })
			address := filepath.Join(dir, "server.sock")
			m := factory(t, "unix", address, nil)
			defer managedStop(t, m)
			if err := m.Listen(); err == nil {
				t.Fatal("expected endpoint extraction error")
			}
			_ = testListener(t, "unix", address).Close()
			if err := m.Start(context.Background()); err == nil {
				t.Fatal("expected Start error")
			}
			_ = testListener(t, "unix", address).Close()
		})
	}
}

type acceptingListener struct {
	net.Listener
	entered chan struct{}
	once    sync.Once
}

func (l *acceptingListener) Accept() (net.Conn, error) {
	l.once.Do(func() { close(l.entered) })
	return l.Listener.Accept()
}

type failingLifecycleListener struct {
	net.Listener
	acceptErr error
	closeErr  error
}

func (l *failingLifecycleListener) Accept() (net.Conn, error) { return nil, l.acceptErr }

func (l *failingLifecycleListener) Close() error {
	_ = l.Listener.Close()
	return l.closeErr
}

func TestManagedPreservesServeAndCleanupErrors(t *testing.T) {
	for name, factory := range managedFactories {
		t.Run(name, func(t *testing.T) {
			raw := testListener(t, "tcp", "127.0.0.1:0")
			acceptErr, closeErr := errors.New("accept failed"), errors.New("listener cleanup failed")
			lis := &failingLifecycleListener{Listener: raw, acceptErr: acceptErr, closeErr: closeErr}
			m := factory(t, "tcp", "127.0.0.1:0", lis)
			err := m.Start(context.Background())
			if !errors.Is(err, acceptErr) || !errors.Is(err, closeErr) {
				t.Fatalf("Start lost serve or cleanup error: %v", err)
			}
			if err := m.Stop(context.Background()); !errors.Is(err, closeErr) {
				t.Fatalf("Stop lost shared cleanup error: %v", err)
			}
			_ = testListener(t, "tcp", raw.Addr().String()).Close()
		})
	}
}

func TestManagedConcurrentStartStop(t *testing.T) {
	for name, factory := range managedFactories {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < 20; i++ {
				lis := &acceptingListener{Listener: testListener(t, "tcp", "127.0.0.1:0"), entered: make(chan struct{})}
				m := factory(t, "tcp", "127.0.0.1:0", lis)
				ctx, cancel := context.WithCancel(context.Background())
				start := make(chan error, 1)
				go func() { start <- m.Start(ctx) }()
				if i%2 == 0 {
					select {
					case <-lis.entered:
					case <-time.After(3 * time.Second):
						t.Fatal("Serve did not start")
					}
					if err := m.Start(ctx); !errors.Is(err, transport.ErrAlreadyStarted) {
						t.Fatalf("second Start: %v", err)
					}
				}
				cancel()
				var wg sync.WaitGroup
				for j := 0; j < 5; j++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						m.EndpointURL()
						managedStop(t, m)
						m.EndpointURL()
					}()
				}
				wg.Wait()
				select {
				case err := <-start:
					if err != nil && !errors.Is(err, transport.ErrServerStopped) && !errors.Is(err, context.Canceled) {
						t.Fatalf("Start during Stop: %v", err)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("Start did not return")
				}
				_ = testListener(t, "tcp", lis.Addr().String()).Close()
			}
		})
	}
}
