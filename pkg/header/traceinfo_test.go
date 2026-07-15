package header

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTraceInfoFromContextPriority(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "ctx-trace")
	ctx = WithSpanID(ctx, "ctx-span")
	ti := TraceInfoFromContext(ctx)
	if ti.TraceID != "ctx-trace" || ti.SpanID != "ctx-span" {
		t.Fatalf("traceinfo = %+v", ti)
	}
}

func TestTraceInfoFromRequestContextOverHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(TraceparentHeaderKey, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	ctx := WithTraceID(req.Context(), "from-ctx")
	req = req.WithContext(ctx)

	ti := TraceInfoFromRequest(req)
	if ti.TraceID != "from-ctx" {
		t.Fatalf("trace_id = %q, want from-ctx", ti.TraceID)
	}
}

func TestTraceInfoFromRequestTraceparent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(TraceparentHeaderKey, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	ti := TraceInfoFromRequest(req)
	if ti.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace_id = %q", ti.TraceID)
	}
	if ti.SpanID != "00f067aa0ba902b7" {
		t.Fatalf("span_id = %q", ti.SpanID)
	}
}

func TestTraceInfoFromRequestXTraceID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(TraceIDKey, "legacy-trace")

	ti := TraceInfoFromRequest(req)
	if ti.TraceID != "legacy-trace" {
		t.Fatalf("trace_id = %q", ti.TraceID)
	}
}

func TestLogTraceIDFallbackRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDKey, "req-only")
	if got := LogTraceID(req); got != "req-only" {
		t.Fatalf("log trace = %q", got)
	}
}

func TestTraceInfoFromHeadersTraceparentOverXTraceID(t *testing.T) {
	h := http.Header{}
	h.Set(TraceparentHeaderKey, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	h.Set(TraceIDKey, "legacy-trace")

	ti := TraceInfoFromHeaders(h)
	if ti.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace_id = %q", ti.TraceID)
	}
}
