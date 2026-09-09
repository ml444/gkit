package response

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/pkg/header"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestWrapResponse_IncludesRequestID(t *testing.T) {
	ctx := header.WithRequestID(context.Background(), "req-service")
	handler := WrapResponse()(func(context.Context, any) (any, error) {
		return wrapperspb.String("test"), nil
	})

	rsp, err := handler(ctx, nil)
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	wrapped, ok := rsp.(*ApiCommonResponse)
	if !ok {
		t.Fatalf("response type = %T", rsp)
	}
	if got := wrapped.GetRequestId(); got != "req-service" {
		t.Fatalf("request_id = %q, want %q", got, "req-service")
	}
}

func TestWrapResponse_GeneratesRequestID(t *testing.T) {
	handler := WrapResponse()(func(context.Context, any) (any, error) {
		return wrapperspb.String("test"), nil
	})

	rsp, err := handler(context.Background(), nil)
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	wrapped, ok := rsp.(*ApiCommonResponse)
	if !ok {
		t.Fatalf("response type = %T", rsp)
	}
	assertRandomRequestID(t, wrapped.GetRequestId())
}

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
	req = req.WithContext(header.WithRequestID(req.Context(), "req-http"))
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
	if got := wrapped["requestId"]; got != "req-http" {
		t.Fatalf("requestId = %v, want %q", got, "req-http")
	}
}

func TestWrapHttpResponse_RequestIDFallsBackToHeader(t *testing.T) {
	handler := WrapHttpResponse()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set(header.RequestIDKey, "req-header")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var wrapped map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := wrapped["requestId"]; got != "req-header" {
		t.Fatalf("requestId = %v, want %q", got, "req-header")
	}
}

func TestWrapHttpResponse_GeneratesRequestID(t *testing.T) {
	handler := WrapHttpResponse()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{}`))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var wrapped map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	requestID, ok := wrapped["requestId"].(string)
	if !ok {
		t.Fatalf("requestId type = %T", wrapped["requestId"])
	}
	assertRandomRequestID(t, requestID)
	if got := rec.Header().Get(header.RequestIDKey); got != requestID {
		t.Fatalf("response X-Request-ID = %q, want %q", got, requestID)
	}
}

func assertRandomRequestID(t *testing.T, requestID string) {
	t.Helper()
	b, err := hex.DecodeString(requestID)
	if err != nil {
		t.Fatalf("request ID %q is not hexadecimal: %v", requestID, err)
	}
	if len(b) != 16 {
		t.Fatalf("request ID byte length = %d, want 16", len(b))
	}
}
