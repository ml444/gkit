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
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Server applies a per-RPC context timeout.
func Server(d time.Duration) middleware.Middleware {
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx, cancel := context.WithTimeout(ctx, d)
			defer cancel()
			return next(ctx, req)
		}
	}
}
