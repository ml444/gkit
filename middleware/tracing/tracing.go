package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"google.golang.org/grpc"

	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
)

// Server returns service middleware that propagates trace ID via context.
func Server() middleware.Middleware {
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if header.TraceIDFromContext(ctx) == "" {
				ctx = header.WithTraceID(ctx, newTraceID())
			}
			ctx = header.PropagateOutgoing(ctx)
			return next(ctx, req)
		}
	}
}

// HTTPMiddleware injects trace ID into request context and response headers.
func HTTPMiddleware() middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ti := header.TraceInfoFromHeaders(r.Header)
			if ti.TraceID == "" {
				ti.TraceID = newTraceID()
			}
			ctx := header.WithTraceID(r.Context(), ti.TraceID)
			if ti.SpanID != "" {
				ctx = header.WithSpanID(ctx, ti.SpanID)
			}
			header.PropagateToResponse(w, ctx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UnaryServerInterceptor is a gRPC server interceptor for trace propagation.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if header.TraceIDFromContext(ctx) == "" {
			ctx = header.WithTraceID(ctx, newTraceID())
		}
		ctx = header.PropagateOutgoing(ctx)
		return handler(ctx, req)
	}
}

func newTraceID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
