package httpx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/transport"
)

func TestTransportAccessors(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/a?b=c", nil)
	req.Header.Set("X-In", "1")
	tr := ClientTransport(req)

	if tr.Kind() != "http" || tr.Endpoint() == "" || tr.Path() != "/a" {
		t.Fatalf("unexpected client transport: %#v", tr)
	}
	httpTr, ok := tr.(*Transport)
	if !ok {
		t.Fatal("expected http transport")
	}
	httpTr.SetEndpoint("endpoint")
	httpTr.SetPath("path")
	httpTr.SetRequestHeader(http.Header{"X-Req": {"r"}})
	httpTr.SetResponseHeader(http.Header{"X-Resp": {"s"}})
	if httpTr.Endpoint() != "endpoint" || httpTr.Path() != "path" || httpTr.PathTemplate() != "/a" {
		t.Fatalf("unexpected fields: %#v", httpTr)
	}
	if httpTr.In().GetFirst("X-Req") != "r" || httpTr.Out().GetFirst("X-Resp") != "s" {
		t.Fatalf("unexpected metadata: in=%#v out=%#v", httpTr.In(), httpTr.Out())
	}
	if httpTr.Request() != req {
		t.Fatal("request mismatch")
	}

	ctx := transport.ToContext(context.Background(), httpTr)
	SetOperation(ctx, "op")
	if httpTr.Path() != "op" {
		t.Fatalf("operation = %q", httpTr.Path())
	}
	if got, ok := GetTransport(ctx); !ok || got != httpTr {
		t.Fatalf("GetTransport = %#v %v", got, ok)
	}
	if got, ok := GetTransport(context.Background()); ok || got != nil {
		t.Fatalf("empty GetTransport = %#v %v", got, ok)
	}
}

func TestRedirect(t *testing.T) {
	r := NewRedirect("/next", http.StatusFound)
	url, code := r.Redirect()
	if url != "/next" || code != http.StatusFound {
		t.Fatalf("redirect = %q %d", url, code)
	}
}

func TestRegisterHealthAndPprof(t *testing.T) {
	var nilServer *Server
	nilServer.RegisterHealth("")
	nilServer.RegisterPprof()

	srv := NewServer()
	srv.RegisterHealth("")
	srv.RegisterPprof()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health code = %d", rec.Code)
	}
}
