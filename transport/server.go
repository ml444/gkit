package transport

import (
	"context"
	"errors"
	"net/url"
	"time"
)

var (
	// ErrLifecycleOwned indicates duplicate management or use of a legacy
	// lifecycle method after ownership has passed to a ManagedServer.
	ErrLifecycleOwned = errors.New("transport: server lifecycle already owned")
	// ErrAlreadyStarted indicates a concurrent Start on a running adapter.
	ErrAlreadyStarted = errors.New("transport: server already started")
	// ErrServerStopped indicates an attempt to bind or start after shutdown began.
	ErrServerStopped = errors.New("transport: server is stopping or stopped")
)

// DefaultShutdownTimeout is the grace period used by zero LifecycleOptions.
const DefaultShutdownTimeout = 10 * time.Second

// LifecycleOptions configures the shared HTTP/gRPC lifecycle adapter.
type LifecycleOptions struct {
	// ShutdownTimeout bounds graceful shutdown before connections are forced
	// closed. Zero uses DefaultShutdownTimeout; negative values are invalid.
	ShutdownTimeout time.Duration
}

// ManagedServer owns a single server's lifecycle. Configure routes, services
// and options before creating it, and use only this interface for lifecycle
// operations afterwards. Its methods may be called concurrently.
type ManagedServer interface {
	// Listen explicitly binds the listener. It is idempotent while bound;
	// a failed Listen can be retried before Start or Stop.
	Listen() error
	// EndpointURL returns a copy without allocating resources. The boolean
	// reports bound/serving state, not application readiness. After shutdown
	// the last address may be returned with false.
	EndpointURL() (*url.URL, bool)
	// Start binds if needed, then blocks until serving and shutdown finish.
	// Cancellation triggers an independent graceful shutdown. Successful
	// shutdown returns nil; serve errors and shutdown timeouts are preserved.
	// An already-canceled context returns its error without starting. Once
	// a Start attempt begins, the server cannot be started again, even on error.
	Start(context.Context) error
	// Stop initiates one shared shutdown, including before Start. The context
	// limits only this caller's wait; LifecycleOptions controls shutdown itself.
	// Further calls wait for the same result. A stopped server cannot restart.
	Stop(context.Context) error
}
