package retry

import (
	"context"
	"errors"
	"time"

	"github.com/ml444/gkit/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Options configures retry behavior for client calls.
type Options struct {
	MaxAttempts int
	Backoff     time.Duration
	RetryIf     func(error) bool
}

// Client wraps a service handler with retries (for httpx/grpcx clients).
func Client(opt Options, next middleware.ServiceHandler) middleware.ServiceHandler {
	if opt.MaxAttempts <= 0 {
		opt.MaxAttempts = 3
	}
	if opt.Backoff <= 0 {
		opt.Backoff = 100 * time.Millisecond
	}
	if opt.RetryIf == nil {
		opt.RetryIf = DefaultRetryable // 只重试安全错误
	}
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		var lastErr error
		for attempt := 0; attempt < opt.MaxAttempts; attempt++ {
			rsp, err := next(ctx, req)
			if err == nil || !opt.RetryIf(err) {
				return rsp, err
			}
			lastErr = err
			if attempt+1 < opt.MaxAttempts {
				time.Sleep(opt.Backoff * time.Duration(attempt+1))
			}
		}
		return nil, lastErr
	}
}

// Middleware returns client retry as chainable middleware.
func Middleware(opt Options) middleware.Middleware {
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return Client(opt, next)
	}
}

// DefaultRetryable 仅对瞬时/网络错误返回 true
func DefaultRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	}
	return false
}
