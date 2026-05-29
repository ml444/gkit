package circuitbreaker

import (
	"context"
	"sync"
	"time"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

var ErrOpen = errorx.CreateError(503, 50301, "CIRCUIT: breaker open")

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

type breaker struct {
	mu            sync.Mutex
	st            state
	failures      int
	threshold     int
	openDuration  time.Duration
	openedAt      time.Time
}

func newBreaker(threshold int, open time.Duration) *breaker {
	return &breaker{threshold: threshold, openDuration: open, st: stateClosed}
}

func (b *breaker) allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	switch b.st {
	case stateOpen:
		if time.Since(b.openedAt) >= b.openDuration {
			b.st = stateHalfOpen
			return true
		}
		return false
	default:
		return true
	}
}

func (b *breaker) record(success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if success {
		b.failures = 0
		b.st = stateClosed
		return
	}
	b.failures++
	if b.failures >= b.threshold {
		b.st = stateOpen
		b.openedAt = time.Now()
	}
}

// Options configures circuit breaker middleware.
type Options struct {
	Threshold    int
	OpenDuration time.Duration
}

// Server returns per-path circuit breaker middleware.
func Server(opt Options) middleware.Middleware {
	if opt.Threshold <= 0 {
		opt.Threshold = 5
	}
	if opt.OpenDuration <= 0 {
		opt.OpenDuration = 30 * time.Second
	}
	breakers := make(map[string]*breaker)
	var mu sync.Mutex
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			key := "default"
			if tr, ok := transport.FromContext(ctx); ok {
				key = tr.Path()
			}
			mu.Lock()
			b, ok := breakers[key]
			if !ok {
				b = newBreaker(opt.Threshold, opt.OpenDuration)
				breakers[key] = b
			}
			mu.Unlock()
			if !b.allow() {
				return nil, ErrOpen
			}
			rsp, err := next(ctx, req)
			b.record(err == nil)
			return rsp, err
		}
	}
}
