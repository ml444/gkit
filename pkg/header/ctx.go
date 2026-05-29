package header

import "context"

type traceIDKey struct{}
type requestIDKey struct{}

// WithTraceID stores trace ID in context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// GetTraceID returns trace ID from context only.
func GetTraceID(ctx context.Context) string {
	if v, ok := ctx.Value(traceIDKey{}).(string); ok {
		return v
	}
	return ""
}

// WithRequestID stores request ID in context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// GetRequestID returns request ID from context only.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey{}).(string); ok {
		return v
	}
	return ""
}

// CorrelationID returns trace ID if present, otherwise request ID.
// Falls back to transport incoming metadata when context values are empty.
func CorrelationID(ctx context.Context) string {
	if id := GetTraceID(ctx); id != "" {
		return id
	}
	if id := GetRequestID(ctx); id != "" {
		return id
	}
	if id := TraceIDFromContext(ctx); id != "" {
		return id
	}
	return RequestIDFromContext(ctx)
}
