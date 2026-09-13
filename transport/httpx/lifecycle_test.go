package httpx

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/ml444/gkit/transport"
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
	s := NewServer(Address("127.0.0.1:0"))
	if _, err := Managed(s, transport.LifecycleOptions{ShutdownTimeout: -1}); err == nil {
		t.Fatal("accepted negative timeout")
	}
	// A listener allocated through the old Endpoint can be adopted.
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
	if err := s.Start(context.Background()); !errors.Is(err, transport.ErrLifecycleOwned) {
		t.Fatalf("legacy Start after Managed: %v", err)
	}
	if err := s.Stop(context.Background()); !errors.Is(err, transport.ErrLifecycleOwned) {
		t.Fatalf("legacy Stop after Managed: %v", err)
	}
	if err := m.Listen(); err != nil {
		t.Fatal(err)
	}
	u, ok := m.EndpointURL()
	if !ok || u.String() != ep.String() {
		t.Fatalf("adopted endpoint: %v, %v", u, ok)
	}
	s = NewServer()
	if err := s.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := Managed(s, transport.LifecycleOptions{}); !errors.Is(err, transport.ErrLifecycleOwned) {
		t.Fatalf("Managed after legacy Stop: %v", err)
	}
}

func TestManagedHTTPGraceAndForcedClose(t *testing.T) {
	for _, force := range []bool{false, true} {
		name := "graceful"
		timeout := 2 * time.Second
		if force {
			name, timeout = "forced", 80*time.Millisecond
		}
		t.Run(name, func(t *testing.T) {
			s := NewServer(Address("127.0.0.1:0"), Timeout(0))
			entered := make(chan context.Context, 1)
			release, handlerDone, shutdown := make(chan struct{}), make(chan struct{}), make(chan struct{})
			s.RegisterOnShutdown(func() { close(shutdown) })
			s.GetRouter().GET("/work", func(w http.ResponseWriter, r *http.Request) {
				defer close(handlerDone)
				entered <- r.Context()
				select {
				case <-release:
					_, _ = io.WriteString(w, "completed")
				case <-r.Context().Done():
				}
			})
			m, err := Managed(s, transport.LifecycleOptions{ShutdownTimeout: timeout})
			if err != nil {
				t.Fatal(err)
			}
			defer m.Stop(context.Background())
			if err := m.Listen(); err != nil {
				t.Fatal(err)
			}
			u, _ := m.EndpointURL()
			type key struct{}
			ctx, cancel := context.WithCancel(context.WithValue(context.Background(), key{}, "request-value"))
			defer cancel()
			start := make(chan error, 1)
			go func() { start <- m.Start(ctx) }()
			clientResult := make(chan error, 1)
			go func() {
				client := &http.Client{Timeout: 4 * time.Second}
				res, err := client.Get(u.String() + "/work")
				if err == nil {
					defer res.Body.Close()
					var body []byte
					body, err = io.ReadAll(res.Body)
					if err == nil && string(body) != "completed" {
						err = errors.New("request did not complete")
					}
				}
				clientResult <- err
			}()
			requestCtx := lifecycleResult(t, entered)
			if requestCtx.Value(key{}) != "request-value" {
				t.Fatal("startup context value was lost")
			}
			cancel()
			lifecycleResult(t, shutdown)
			if !force {
				if err := requestCtx.Err(); err != nil {
					t.Fatalf("startup cancellation canceled active request: %v", err)
				}
				select {
				case err := <-start:
					t.Fatalf("Start returned with active request: %v", err)
				default:
				}
				close(release)
				if err := lifecycleResult(t, clientResult); err != nil {
					t.Fatalf("request: %v", err)
				}
				if err := lifecycleResult(t, start); err != nil {
					t.Fatalf("graceful Start: %v", err)
				}
			} else {
				if err := lifecycleResult(t, start); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("forced Start: %v", err)
				}
				lifecycleResult(t, requestCtx.Done())
				if err := lifecycleResult(t, clientResult); err == nil {
					t.Fatal("request unexpectedly completed before forced close")
				}
				if err := m.Stop(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("repeated Stop lost timeout: %v", err)
				}
			}
			lifecycleResult(t, handlerDone)
		})
	}
}

func TestManagedHTTPSEndpointAndStartFailure(t *testing.T) {
	s := NewServer(Address("127.0.0.1:0"), TLSConfig(&tls.Config{}))
	m, err := Managed(s, transport.LifecycleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Stop(context.Background())
	if err := m.Listen(); err != nil {
		t.Fatal(err)
	}
	u, ok := m.EndpointURL()
	if !ok || u.Scheme != "https" {
		t.Fatalf("TLS endpoint: %v, %v", u, ok)
	}
	if err := m.Start(context.Background()); err == nil {
		t.Fatal("missing TLS certificate did not fail Start")
	}
	lis, err := net.Listen("tcp", u.Host)
	if err != nil {
		t.Fatalf("failed Start leaked listener: %v", err)
	}
	_ = lis.Close()
	if _, ok := m.EndpointURL(); ok {
		t.Fatal("failed server is still bound")
	}
}

func TestManagedExplicitEndpointIsDetached(t *testing.T) {
	advertised := &url.URL{Scheme: "https", Host: "api.example.test", Path: "/v1"}
	s := NewServer(Address("127.0.0.1:0"), Endpoint(advertised))
	m, err := Managed(s, transport.LifecycleOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer m.Stop(context.Background())
	advertised.Host = "changed"
	if err := m.Listen(); err != nil {
		t.Fatal(err)
	}
	u, ok := m.EndpointURL()
	if !ok || u.String() != "https://api.example.test/v1" {
		t.Fatalf("advertised URL: %v, %v", u, ok)
	}
}
