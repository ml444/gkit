package lifecycle

import (
	"context"
	"errors"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ml444/gkit/transport"
)

func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(3 * time.Second):
		t.Fatal("lifecycle operation did not finish")
		var zero T
		return zero
	}
}

func TestListenRetryAndConcurrentSnapshots(t *testing.T) {
	var calls atomic.Int32
	bindErr := errors.New("bind failed")
	entered, release := make(chan struct{}), make(chan struct{})
	var c *Controller
	c, _ = New(transport.LifecycleOptions{}, Hooks{
		Bind: func(context.Context) (*url.URL, error) {
			// Calling back into the controller must not deadlock.
			c.EndpointURL()
			if calls.Add(1) == 1 {
				return nil, bindErr
			}
			close(entered)
			<-release
			return &url.URL{Scheme: "http", Host: "localhost:1234"}, nil
		},
		Shutdown: func(context.Context) error { c.EndpointURL(); return nil },
	})
	if err := c.Listen(); !errors.Is(err, bindErr) {
		t.Fatalf("first Listen: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Listen(); err != nil {
				t.Errorf("Listen: %v", err)
			}
		}()
	}
	receive(t, entered)
	if u, ok := c.EndpointURL(); u != nil || ok {
		t.Fatalf("binding endpoint: %v, %v", u, ok)
	}
	close(release)
	wg.Wait()
	if calls.Load() != 2 {
		t.Fatalf("Bind calls = %d", calls.Load())
	}
	u, ok := c.EndpointURL()
	if !ok || u.Host != "localhost:1234" {
		t.Fatalf("bound endpoint: %v, %v", u, ok)
	}
	u.Host = "mutated"
	v, _ := c.EndpointURL()
	if v.Host != "localhost:1234" {
		t.Fatal("EndpointURL returned shared memory")
	}
	if err := c.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if u, ok := c.EndpointURL(); u == nil || ok {
		t.Fatalf("stopped endpoint: %v, %v", u, ok)
	}
}

func TestSharedStopWaitsForShutdown(t *testing.T) {
	entered, serveStop := make(chan struct{}), make(chan struct{})
	shutdownEntered, shutdownRelease := make(chan struct{}), make(chan struct{})
	normal := errors.New("normal stop")
	var shutdowns atomic.Int32
	c, _ := New(transport.LifecycleOptions{}, Hooks{
		Bind: func(context.Context) (*url.URL, error) { return &url.URL{Host: "localhost:1234"}, nil },
		Serve: func(context.Context) error {
			close(entered)
			<-serveStop
			return normal
		},
		Shutdown: func(ctx context.Context) error {
			shutdowns.Add(1)
			close(serveStop)
			close(shutdownEntered)
			<-shutdownRelease
			return ctx.Err()
		},
		IsNormalStop: func(err error) bool { return errors.Is(err, normal) },
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startResult := make(chan error, 1)
	go func() { startResult <- c.Start(ctx) }()
	receive(t, entered)
	if err := c.Start(context.Background()); !errors.Is(err, transport.ErrAlreadyStarted) {
		t.Fatalf("concurrent Start: %v", err)
	}
	caller, cancelCaller := context.WithCancel(context.Background())
	cancelCaller()
	if err := c.Stop(caller); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Stop waiter: %v", err)
	}
	receive(t, shutdownEntered)
	cancel()
	if _, ok := c.EndpointURL(); ok {
		t.Fatal("stopping server reported bound")
	}
	select {
	case err := <-startResult:
		t.Fatalf("Start returned before shutdown finished: %v", err)
	default:
	}
	close(shutdownRelease)
	if err := receive(t, startResult); err != nil {
		t.Fatalf("Start: %v", err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Stop(context.Background()); err != nil {
				t.Errorf("repeated Stop: %v", err)
			}
		}()
	}
	wg.Wait()
	if shutdowns.Load() != 1 {
		t.Fatalf("shutdown count: %d", shutdowns.Load())
	}
	if err := c.Listen(); !errors.Is(err, transport.ErrServerStopped) {
		t.Fatalf("Listen after Stop: %v", err)
	}
	if err := c.Start(context.Background()); !errors.Is(err, transport.ErrServerStopped) {
		t.Fatalf("Start after Stop: %v", err)
	}
}

func TestCancelStartDuringBind(t *testing.T) {
	entered := make(chan struct{})
	var cleaned atomic.Bool
	c, _ := New(transport.LifecycleOptions{}, Hooks{
		Bind: func(ctx context.Context) (*url.URL, error) {
			close(entered)
			<-ctx.Done()
			cleaned.Store(true)
			return nil, ctx.Err()
		},
		Shutdown: func(context.Context) error {
			if !cleaned.Load() {
				t.Error("shutdown raced unfinished bind")
			}
			return nil
		},
		IsNormalStop: func(error) bool { return false },
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- c.Start(ctx) }()
	receive(t, entered)
	cancel()
	if err := receive(t, done); err != nil {
		t.Fatalf("cancel during bind: %v", err)
	}
}

func TestStopWaitsForSuccessfulInFlightBind(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	var cleaned atomic.Bool
	c, _ := New(transport.LifecycleOptions{}, Hooks{
		Bind: func(context.Context) (*url.URL, error) {
			close(entered)
			<-release
			return &url.URL{Host: "localhost:1234"}, nil
		},
		Shutdown: func(context.Context) error { cleaned.Store(true); return nil },
	})
	listen := make(chan error, 1)
	go func() { listen <- c.Listen() }()
	receive(t, entered)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Stop(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Stop while binding: %v", err)
	}
	if cleaned.Load() {
		t.Fatal("shutdown ran before listener publication")
	}
	close(release)
	if err := receive(t, listen); !errors.Is(err, transport.ErrServerStopped) {
		t.Fatalf("Listen raced Stop: %v", err)
	}
	if err := c.Stop(context.Background()); err != nil || !cleaned.Load() {
		t.Fatalf("Stop: %v, cleaned=%v", err, cleaned.Load())
	}
	if u, ok := c.EndpointURL(); u == nil || ok {
		t.Fatalf("late bind resurrected server: %v, %v", u, ok)
	}
}

func TestStartErrorAndTimeoutAreBothPreserved(t *testing.T) {
	serveErr := errors.New("unexpected serve failure")
	c, _ := New(transport.LifecycleOptions{ShutdownTimeout: time.Millisecond}, Hooks{
		Bind:         func(context.Context) (*url.URL, error) { return &url.URL{}, nil },
		Serve:        func(context.Context) error { return serveErr },
		Shutdown:     func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
		IsNormalStop: func(err error) bool { return errors.Is(err, serveErr) },
	})
	err := c.Start(context.Background())
	if !errors.Is(err, serveErr) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Start lost an error: %v", err)
	}
	if err := c.Stop(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Stop lost shutdown result: %v", err)
	}
}

func TestInitiallyCanceledStartDoesNotConsumeServer(t *testing.T) {
	var binds int
	c, _ := New(transport.LifecycleOptions{}, Hooks{
		Bind:     func(context.Context) (*url.URL, error) { binds++; return &url.URL{}, nil },
		Shutdown: func(context.Context) error { return nil },
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := c.Start(ctx); !errors.Is(err, context.Canceled) || binds != 0 {
		t.Fatalf("canceled Start: %v, binds=%d", err, binds)
	}
	if err := c.Start(nil); err == nil {
		t.Fatal("accepted nil context")
	}
	if err := c.Stop(nil); err == nil {
		t.Fatal("accepted nil context")
	}
	if err := c.Listen(); err != nil {
		t.Fatal(err)
	}
	if err := c.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if c.timeout != transport.DefaultShutdownTimeout {
		t.Fatalf("default timeout = %s", c.timeout)
	}
	if _, err := New(transport.LifecycleOptions{ShutdownTimeout: -1}, Hooks{}); err == nil {
		t.Fatal("accepted negative timeout")
	}
}
