package header

import (
	"context"

	"github.com/ml444/gkit/transport"
)

// TraceIDFromContext resolves trace ID from context values, then incoming transport metadata.
func TraceIDFromContext(ctx context.Context) string {
	if id := GetTraceID(ctx); id != "" {
		return id
	}
	return firstIncoming(ctx, traceKeyVariants...)
}

// RequestIDFromContext resolves request ID from context values, then incoming transport metadata.
func RequestIDFromContext(ctx context.Context) string {
	if id := GetRequestID(ctx); id != "" {
		return id
	}
	return firstIncoming(ctx, requestIDKeyVariants...)
}

// SetIncoming attaches key/value to incoming transport metadata when present.
func SetIncoming(ctx context.Context, key, value string) context.Context {
	if key == "" || value == "" {
		return ctx
	}
	tr, ok := transport.FromContext(ctx)
	if !ok {
		return ctx
	}
	md := tr.In()
	if md == nil {
		return ctx
	}
	md.Set(key, value)
	return ctx
}

// SetOutgoing attaches key/value to outgoing transport metadata when present.
func SetOutgoing(ctx context.Context, key, value string) context.Context {
	if key == "" || value == "" {
		return ctx
	}
	tr, ok := transport.FromContext(ctx)
	if !ok {
		return ctx
	}
	md := tr.Out()
	if md == nil {
		return ctx
	}
	md.Set(key, value)
	return ctx
}

// PropagateOutgoing copies trace and request IDs from context to outgoing transport metadata.
func PropagateOutgoing(ctx context.Context) context.Context {
	if id := GetTraceID(ctx); id != "" {
		ctx = SetOutgoing(ctx, TraceIDKey, id)
	}
	if id := GetRequestID(ctx); id != "" {
		ctx = SetOutgoing(ctx, RequestIDKey, id)
	}
	return ctx
}

// FirstIncoming returns the first non-empty value for keys from incoming transport metadata.
func FirstIncoming(ctx context.Context, keys ...string) string {
	return firstIncoming(ctx, keys...)
}

func firstIncoming(ctx context.Context, keys ...string) string {
	tr, ok := transport.FromContext(ctx)
	if !ok {
		return ""
	}
	return firstMD(tr.In(), keys...)
}

func firstMD(md transport.MD, keys ...string) string {
	if md == nil {
		return ""
	}
	for _, k := range keys {
		if v := md.GetFirst(k); v != "" {
			return v
		}
	}
	return ""
}
