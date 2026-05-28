package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrapHttpResponse_SkipsRawResponse(t *testing.T) {
	handler := WrapHttpResponse()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		MarkHttpRaw(w)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte{0x01, 0x02, 0x03})
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get(HttpRawHeader) != "1" {
		t.Fatalf("expected raw header, got %q", rec.Header().Get(HttpRawHeader))
	}
	if rec.Header().Get("Content-Type") != "application/octet-stream" {
		t.Fatalf("content-type = %q", rec.Header().Get("Content-Type"))
	}
	if got := rec.Body.Bytes(); string(got) != "\x01\x02\x03" {
		t.Fatalf("body = %q, want raw bytes", got)
	}
	var wrapped map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err == nil {
		t.Fatal("expected raw body, got JSON wrapper")
	}
}

func TestWrapHttpResponse_WrapsJSONResponse(t *testing.T) {
	handler := WrapHttpResponse()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"test"}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var wrapped map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if wrapped["code"].(float64) != 0 {
		t.Fatalf("code = %v", wrapped["code"])
	}
}
