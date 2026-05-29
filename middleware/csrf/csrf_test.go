package csrf

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCSRF_SkipsBearerByDefault(t *testing.T) {
	h := HTTPMiddleware(DefaultOptions())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer token-abc")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for bearer API", rec.Code)
	}
}

func TestCSRF_EnforcesWhenBearerDisabled(t *testing.T) {
	opt := DefaultOptions()
	opt.SkipBearer = false
	h := HTTPMiddleware(opt)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Authorization", "Bearer token-abc")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 when SkipBearer=false", rec.Code)
	}
}
