// Package lifecycle coordinates protocol-specific servers without holding
// state locks across network operations or server callbacks.
package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/ml444/gkit/transport"
)

type Hooks struct {
	Bind         func(context.Context) (*url.URL, error)
	Serve        func(context.Context) error
	Shutdown     func(context.Context) error
	IsNormalStop func(error) bool
}

type phase uint8

const (
	newServer phase = iota
	binding
	bound
	serving
	stopping
	stopped
)

type bindAttempt struct {
	done   chan struct{}
	cancel context.CancelFunc
	err    error // published before done is closed
}

type Controller struct {
	mu        sync.Mutex
	phase     phase
	endpoint  *url.URL
	bind      *bindAttempt
	started   bool
	serveDone chan struct{}
	stopDone  chan struct{}
	stopErr   error // published before stopDone is closed
	timeout   time.Duration
	hooks     Hooks
}

var _ transport.ManagedServer = (*Controller)(nil)

func New(opts transport.LifecycleOptions, hooks Hooks) (*Controller, error) {
	if opts.ShutdownTimeout < 0 {
		return nil, fmt.Errorf("transport: negative shutdown timeout: %s", opts.ShutdownTimeout)
	}
	if opts.ShutdownTimeout == 0 {
		opts.ShutdownTimeout = transport.DefaultShutdownTimeout
	}
	return &Controller{timeout: opts.ShutdownTimeout, hooks: hooks}, nil
}

func CloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	copy := *u
	return &copy
}

// CloseListener also handles a listener that was bound but never served.
func CloseListener(lis net.Listener) error {
	if lis == nil {
		return nil
	}
	err := lis.Close()
	if errors.Is(err, net.ErrClosed) {
		return nil
	}
	return err
}

func (c *Controller) EndpointURL() (*url.URL, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return CloneURL(c.endpoint), c.phase == bound || c.phase == serving
}

func (c *Controller) Listen() error {
	c.mu.Lock()
	switch c.phase {
	case stopping, stopped:
		c.mu.Unlock()
		return transport.ErrServerStopped
	case bound, serving:
		c.mu.Unlock()
		return nil
	case binding:
		attempt := c.bind
		c.mu.Unlock()
		<-attempt.done
		return attempt.err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	attempt := &bindAttempt{done: make(chan struct{}), cancel: cancel}
	c.bind, c.phase = attempt, binding
	c.mu.Unlock()

	endpoint, err := c.hooks.Bind(ctx)
	c.mu.Lock()
	if err == nil {
		c.endpoint = CloneURL(endpoint)
	}
	if c.phase == binding {
		c.phase = newServer
		if err == nil {
			c.phase = bound
		}
	} else if err == nil {
		err = transport.ErrServerStopped
	}
	attempt.err = err
	close(attempt.done)
	c.mu.Unlock()
	return err
}

func (c *Controller) Start(ctx context.Context) error {
	if ctx == nil {
		return errors.New("transport: nil start context")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	if c.phase >= stopping {
		c.mu.Unlock()
		return transport.ErrServerStopped
	}
	if c.started {
		c.mu.Unlock()
		return transport.ErrAlreadyStarted
	}
	c.started = true
	c.serveDone = make(chan struct{})
	c.mu.Unlock()

	finished, watcherDone := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(watcherDone)
		select {
		case <-ctx.Done():
			c.beginStop()
		case <-finished:
		}
	}()
	defer func() { close(finished); <-watcherDone }()

	err := c.Listen()
	if ctx.Err() != nil {
		c.beginStop()
	}
	c.mu.Lock()
	run := err == nil && c.phase < stopping
	if run {
		c.phase = serving
	}
	c.mu.Unlock()
	if run {
		err = c.hooks.Serve(ctx)
	}

	c.mu.Lock()
	expectedStop := c.phase >= stopping
	close(c.serveDone)
	c.mu.Unlock()
	if expectedStop && (c.hooks.IsNormalStop(err) ||
		(!run && (errors.Is(err, context.Canceled) || errors.Is(err, transport.ErrServerStopped)))) {
		err = nil
	}
	<-c.beginStop()
	return errors.Join(err, c.stopErr)
}

func (c *Controller) beginStop() <-chan struct{} {
	c.mu.Lock()
	if c.stopDone != nil {
		done := c.stopDone
		c.mu.Unlock()
		return done
	}
	c.stopDone, c.phase = make(chan struct{}), stopping
	attempt, serveDone, done := c.bind, c.serveDone, c.stopDone
	c.mu.Unlock()
	if attempt != nil {
		attempt.cancel()
	}
	go func() {
		if attempt != nil {
			<-attempt.done
		}
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		err := c.hooks.Shutdown(ctx)
		cancel()
		if serveDone != nil {
			<-serveDone
		}
		c.mu.Lock()
		c.phase, c.stopErr = stopped, err
		close(c.stopDone)
		c.mu.Unlock()
	}()
	return done
}

func (c *Controller) Stop(ctx context.Context) error {
	if ctx == nil {
		return errors.New("transport: nil stop context")
	}
	done := c.beginStop()
	select {
	case <-done:
		return c.stopErr
	default:
	}
	select {
	case <-done:
		return c.stopErr
	case <-ctx.Done():
		return ctx.Err()
	}
}
