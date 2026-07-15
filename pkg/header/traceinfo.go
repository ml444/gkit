package header

import (
	"context"
	"net/http"
)

// TraceInfoFromHeaders resolves trace info from HTTP headers (no request-id fallback).
func TraceInfoFromHeaders(h http.Header) TraceInfo {
	if h == nil {
		return TraceInfo{}
	}
	if tp := h.Get(TraceparentHeaderKey); tp != "" {
		if ti, ok := ParseTraceparent(tp); ok {
			return ti
		}
	}
	if id := firstHeader(h, traceKeyVariants...); id != "" {
		return TraceInfo{TraceID: id}
	}
	return TraceInfo{}
}

// TraceInfoFromContext resolves trace info from context values, then incoming transport metadata.
func TraceInfoFromContext(ctx context.Context) TraceInfo {
	if ctx == nil {
		return TraceInfo{}
	}
	ti := TraceInfo{
		TraceID: GetTraceID(ctx),
		SpanID:  GetSpanID(ctx),
	}
	if ti.TraceID != "" {
		return ti
	}
	if id := firstIncoming(ctx, traceKeyVariants...); id != "" {
		return TraceInfo{TraceID: id}
	}
	return TraceInfo{}
}

// TraceInfoFromRequest resolves trace info with context taking priority over headers.
func TraceInfoFromRequest(r *http.Request) TraceInfo {
	if r == nil {
		return TraceInfo{}
	}
	if ti := TraceInfoFromContext(r.Context()); ti.TraceID != "" {
		return ti
	}
	return TraceInfoFromHeaders(r.Header)
}

// LogTraceID returns trace_id for logging, falling back to request-id when trace is absent.
func LogTraceID(r *http.Request) string {
	if r == nil {
		return ""
	}
	if id := TraceInfoFromRequest(r).TraceID; id != "" {
		return id
	}
	return RequestIDFromRequest(r)
}
