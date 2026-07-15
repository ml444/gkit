package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/pkg/header"
)

func TestHTTPMiddlewareObserveDurationWithTrace(t *testing.T) {
	rec := NewInMemoryRecorder()
	SetRecorder(rec)

	h := HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	ctx := header.WithTraceID(req.Context(), "4bf92f3577b34da6a3ce929d0e0e4736")
	ctx = header.WithSpanID(ctx, "00f067aa0ba902b7")
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if len(rec.Durations) != 1 {
		t.Fatalf("durations = %d, want 1", len(rec.Durations))
	}
}
