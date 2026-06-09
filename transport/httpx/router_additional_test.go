package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/middleware"
)

func TestRouterAllMethodsAndGroups(t *testing.T) {
	r := newRouter("/", NewRouterCfg())
	handler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
	r.HEAD("/head", handler)
	r.POST("/post", handler)
	r.PUT("/put", handler)
	r.PATCH("/patch", handler)
	r.DELETE("/delete", handler)
	r.CONNECT("/connect", handler)
	r.OPTIONS("/options", handler)
	r.TRACE("/trace", handler)
	r.Group("/api").GET("/users", handler)

	tests := []struct {
		method string
		path   string
	}{
		{http.MethodHead, "/head"},
		{http.MethodPost, "/post"},
		{http.MethodPut, "/put"},
		{http.MethodPatch, "/patch"},
		{http.MethodDelete, "/delete"},
		{http.MethodConnect, "/connect"},
		{http.MethodOptions, "/options"},
		{http.MethodTrace, "/trace"},
		{http.MethodGet, "/api/users"},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(tt.method, tt.path, nil)
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("%s %s code = %d", tt.method, tt.path, rec.Code)
		}
	}
}

func TestServerMiddlewaresAndRouteGroup(t *testing.T) {
	s := NewServer()
	mw := func(next http.Handler) http.Handler { return next }
	g := s.NewRouteGroup("/v1", mw)
	if g == nil {
		t.Fatal("expected group")
	}
	s.SetMiddlewares(func(next middleware.ServiceHandler) middleware.ServiceHandler { return next })
	if len(s.Middlewares()) != 1 {
		t.Fatalf("middlewares = %d", len(s.Middlewares()))
	}
}
