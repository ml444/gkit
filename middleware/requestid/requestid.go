package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
)

// FromContext returns the request ID from context.
func FromContext(ctx context.Context) string {
	return header.GetRequestID(ctx)
}

// HTTPMiddleware injects or propagates X-Request-ID.
func HTTPMiddleware() middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := header.RequestIDFromRequest(r)
			if id == "" {
				id = newID()
			}
			header.SetRequestID(w.Header(), id)
			ctx := header.WithRequestID(r.Context(), id)
			if header.GetTraceID(ctx) == "" {
				ctx = header.WithTraceID(ctx, id)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Server returns service middleware that ensures request ID in context.
func Server() middleware.Middleware {
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if header.GetRequestID(ctx) == "" {
				id := newID()
				ctx = header.WithRequestID(ctx, id)
				if header.GetTraceID(ctx) == "" {
					ctx = header.WithTraceID(ctx, id)
				}
			}
			return next(ctx, req)
		}
	}
}

func newID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
