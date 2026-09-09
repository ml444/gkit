package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

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
				id = NewID()
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
				id := NewID()
				ctx = header.WithRequestID(ctx, id)
				if header.GetTraceID(ctx) == "" {
					ctx = header.WithTraceID(ctx, id)
				}
			}
			return next(ctx, req)
		}
	}
}

// NewID returns a new random request ID.
func NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// 安全随机失败时退回纳秒时间戳。
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b[:])
}
