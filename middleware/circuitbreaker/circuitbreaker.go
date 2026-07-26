package circuitbreaker

import (
	"context"
	"sync"
	"time"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/containerx/lru"
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
	mu           sync.Mutex
	st           state
	failures     int
	threshold    int
	openDuration time.Duration
	openedAt     time.Time
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
	case stateHalfOpen:
		return false // 探测期间拒绝其余请求
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
	if b.st == stateHalfOpen { // 探针失败：立即重新打开
		b.st = stateOpen
		b.openedAt = time.Now()
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
	MaxBreakers  int // 新增：允许的最大断路器数量
}

// Server returns per-path circuit breaker middleware.
func Server(opt Options) middleware.Middleware {
	if opt.Threshold <= 0 {
		opt.Threshold = 5
	}
	if opt.OpenDuration <= 0 {
		opt.OpenDuration = 30 * time.Second
	}
	if opt.MaxBreakers <= 0 {
        opt.MaxBreakers = 5000 // 默认最多保留 5000 个路由规则
    }
	cache := lru.NewLRUCache[string, *breaker](opt.MaxBreakers)
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			key := "default"
			if tr, ok := transport.FromContext(ctx); ok {
				key = tr.Path()
			}
			var b *breaker
            // 尝试从 LRU 获取，如果没有则新建
            if val, ok := cache.Get(key); ok {
                b = val
            } else {
                // 注意：高并发下可能会有多个请求同时走到这里并创建新 breaker
                // 对于熔断器来说，偶尔的覆盖是可接受的。如果要求绝对精确，可以加一把细粒度的锁或者使用 golang.org/x/sync/singleflight
                b = newBreaker(opt.Threshold, opt.OpenDuration)
                cache.Put(key, b)
            }
			if !b.allow() {
				return nil, ErrOpen
			}
			rsp, err := next(ctx, req)
			b.record(err == nil)
			return rsp, err
		}
	}
}
