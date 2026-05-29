package header

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/httpx"
)

func TestTraceAndRequestIDContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-1")
	ctx = WithRequestID(ctx, "req-1")
	if got := GetTraceID(ctx); got != "trace-1" {
		t.Fatalf("trace = %q", got)
	}
	if got := GetRequestID(ctx); got != "req-1" {
		t.Fatalf("request = %q", got)
	}
	if got := CorrelationID(ctx); got != "trace-1" {
		t.Fatalf("correlation = %q", got)
	}
}

func TestTraceIDFromRequestFallback(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDKey, "req-only")
	if got := TraceIDFromRequest(req); got != "req-only" {
		t.Fatalf("trace from request = %q", got)
	}
}

func TestClientIPFromHeaders(t *testing.T) {
	h := http.Header{}
	h.Set(HeaderCFConnectingIP, "203.0.113.10")
	ip := ClientIPFromHeaders(h, "127.0.0.1:1234", ClientIPOptions{TrustForwarded: true})
	if ip != "203.0.113.10" {
		t.Fatalf("ip = %q", ip)
	}
}

func TestTraceIDFromTransport(t *testing.T) {
	tr := &httpx.Transport{}
	tr.SetRequestHeader(http.Header{TraceIDKey: []string{"from-md"}})
	ctx := transport.ToContext(context.Background(), tr)
	if got := TraceIDFromContext(ctx); got != "from-md" {
		t.Fatalf("trace from transport = %q", got)
	}
}

func TestPropagateOutgoing(t *testing.T) {
	tr := &httpx.Transport{}
	tr.SetResponseHeader(http.Header{})
	ctx := transport.ToContext(context.Background(), tr)
	ctx = WithTraceID(ctx, "t-99")
	ctx = PropagateOutgoing(ctx)
	if got := tr.Out().GetFirst(TraceIDKey); got != "t-99" {
		t.Fatalf("out trace = %q", got)
	}
}

func TestForwardHeaders(t *testing.T) {
	src := http.Header{}
	src.Set(TraceIDKey, "abc")
	dst := http.Header{}
	Forward(dst, src)
	if dst.Get(TraceIDKey) != "abc" {
		t.Fatalf("forward failed")
	}
}
