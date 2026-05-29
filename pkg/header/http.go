package header

import (
	"context"
	"net/http"
)

// Get returns the first header value for key (HTTP canonical lookup).
func Get(h http.Header, key string) string {
	if h == nil {
		return ""
	}
	return h.Get(key)
}

// SetTraceID writes trace ID to HTTP headers.
func SetTraceID(h http.Header, traceID string) {
	if traceID != "" {
		h.Set(TraceIDKey, traceID)
	}
}

// SetRequestID writes request ID to HTTP headers.
func SetRequestID(h http.Header, requestID string) {
	if requestID != "" {
		h.Set(RequestIDKey, requestID)
	}
}

// PropagateToResponse writes trace and request IDs from ctx onto the response.
func PropagateToResponse(w http.ResponseWriter, ctx context.Context) {
	if id := GetTraceID(ctx); id != "" {
		SetTraceID(w.Header(), id)
	}
	if id := GetRequestID(ctx); id != "" {
		SetRequestID(w.Header(), id)
	}
}

func TraceIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	if id := firstHeader(r.Header, traceKeyVariants...); id != "" {
		return id
	}
	return RequestIDFromRequest(r)
}

// RequestIDFromRequest reads request ID from headers.
func RequestIDFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	return firstHeader(r.Header, requestIDKeyVariants...)
}

// Forward copies propagation headers from src to dst when present.
func Forward(dst, src http.Header) {
	if dst == nil || src == nil {
		return
	}
	for _, key := range []string{TraceIDKey, RequestIDKey, RemoteIPKey, ClientIDKey, ClientTypeKey} {
		if v := src.Get(key); v != "" {
			dst.Set(key, v)
		}
	}
}

func firstHeader(h http.Header, keys ...string) string {
	for _, k := range keys {
		if v := h.Get(k); v != "" {
			return v
		}
	}
	return ""
}
