package timeout

import (
	"context"
	"net/http"
	"time"

	"github.com/ml444/gkit/middleware"
)

// HTTPMiddleware applies a per-request context timeout.
func HTTPMiddleware(d time.Duration) middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, d, "Request Timeout")
	}
}

// Server applies a per-RPC context timeout.
func Server(d time.Duration) middleware.Middleware {
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx, cancel := context.WithTimeout(ctx, d)
            defer cancel()

            type result struct {
                res interface{}
                err error
            }
            ch := make(chan result, 1)

            go func() {
                res, err := next(ctx, req)
                ch <- result{res: res, err: err}
            }()

            select {
            case <-ctx.Done():
                return nil, ctx.Err() // Returns context.DeadlineExceeded immediately
            case res := <-ch:
                return res.res, res.err
            }
		}
	}
}
