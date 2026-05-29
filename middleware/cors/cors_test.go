package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS_Preflight(t *testing.T) {
	h := New(Options{AllowOrigins: []string{"https://example.com"}, MaxAge: 600})(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatalf("origin = %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
