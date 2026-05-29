package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/pkg/header"
)

func TestHTTPMiddleware_SetsRequestID(t *testing.T) {
	var got string
	h := HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = FromContext(r.Context())
	}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)
	if got == "" {
		t.Fatal("expected request id in context")
	}
	if rec.Header().Get(header.RequestIDKey) == "" {
		t.Fatal("expected request id header")
	}
}
