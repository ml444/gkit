package tracing

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/pkg/header"
)

func TestHTTPMiddlewareTraceparent(t *testing.T) {
	var got header.TraceInfo
	h := HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = header.TraceInfoFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set(header.TraceparentHeaderKey, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got.TraceID != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("trace_id = %q", got.TraceID)
	}
	if got.SpanID != "00f067aa0ba902b7" {
		t.Fatalf("span_id = %q", got.SpanID)
	}
	if rr.Header().Get(header.TraceIDKey) != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Fatalf("response trace header = %q", rr.Header().Get(header.TraceIDKey))
	}
}

func TestHTTPMiddlewareInvalidTraceparentGeneratesTrace(t *testing.T) {
	var got string
	h := HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = header.GetTraceID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(header.TraceparentHeaderKey, "invalid")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if len(got) != 32 {
		t.Fatalf("expected generated 32-hex trace_id, got %q", got)
	}
}
