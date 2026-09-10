package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestClientURLAndTLS(t *testing.T) {
	s := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			t.Error("request was not TLS")
		}
		if r.URL.EscapedPath() != "/base%2Fsegment/items/a%2Fb" || r.URL.Query().Get("q") != "x y" {
			t.Errorf("URL = %s", r.URL)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	defer s.Close()
	c, err := NewClient(WithEndpoint(s.URL+"/base%2Fsegment"), WithTransport(s.Client().Transport))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err = c.Invoke(context.Background(), "GET", "/items/a%2Fb?q=x+y", nil, nil); err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("GET", "http://unused/items/a%2Fb?q=x+y", nil)
	original := req.URL.String()
	for i := 0; i < 2; i++ {
		res, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
	}
	if req.URL.String() != original || req.Host != "unused" {
		t.Fatalf("caller request mutated: %s %s", req.URL, req.Host)
	}
	untrusted, err := NewClient(WithEndpoint(s.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer untrusted.Close()
	if err = untrusted.Invoke(context.Background(), "GET", "/", nil, nil); err == nil {
		t.Fatal("untrusted server certificate accepted")
	}
}

func TestClientConfigurationValidation(t *testing.T) {
	for _, endpoint := range []string{"discovery:///", "discovery://svc", "discovery:///svc", "ftp://host", "http://", "http://u:p@host", "http://host?q=1", "http://host#", "http://host:99999"} {
		t.Run(endpoint, func(t *testing.T) {
			if _, err := NewClient(WithEndpoint(endpoint)); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
	if _, err := NewClient(WithTimeout(-time.Second)); err == nil {
		t.Fatal("negative timeout accepted")
	}
	c, err := NewClient()
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Invoke(context.Background(), "GET", "/", nil, nil); err == nil {
		t.Fatal("Invoke accepted missing endpoint")
	}
	if _, err = c.Do(nil); err == nil {
		t.Fatal("nil request accepted")
	}
}

func TestDiscoveryIPv6(t *testing.T) {
	reg := discovery.NewDefaultRegistry()
	if err := reg.Register(context.Background(), &discovery.ServiceInstance{ID: "v6", Name: "svc", Address: "::1", Port: 8080}); err != nil {
		t.Fatal(err)
	}
	c, err := NewClient(WithDiscovery(discovery.NewDiscoveryClient(reg), "svc"), WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Host != "[::1]:8080" || r.Host != r.URL.Host {
			t.Errorf("host = %q / %q", r.URL.Host, r.Host)
		}
		return &http.Response{StatusCode: 204, Header: make(http.Header), Body: http.NoBody}, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	if err = c.Invoke(context.Background(), "GET", "/", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestMiddlewareHeaderIsolation(t *testing.T) {
	shared := http.Header{"Authorization": {"old"}, "X-Delete": {"remove"}}
	original := shared.Clone()
	c, err := NewClient(WithEndpoint("example.test"), WithMiddlewares(func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, in interface{}) (interface{}, error) {
			tr, _ := transport.FromContext(ctx)
			tr.In().Set("Authorization", tr.Path())
			tr.In().Delete("X-Delete")
			return next(ctx, in)
		}
	}), WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Header.Get("Authorization") != r.URL.Path || r.Header.Get("X-Delete") != "" {
			t.Errorf("outbound headers %v for %s", r.Header, r.URL)
		}
		return &http.Response{StatusCode: 204, Header: make(http.Header), Body: http.NoBody}, nil
	})))
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, path := range []string{"/one", "/two", "/three"} {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			if err := c.Invoke(context.Background(), "GET", path, nil, nil, RequestHeader(shared), AddRequestHeader("X-Added", "1")); err != nil {
				t.Error(err)
			}
		}(path)
	}
	wg.Wait()
	if !reflect.DeepEqual(shared, original) {
		t.Fatalf("caller headers changed: %v", shared)
	}
}

func TestGroupAllRegistrationMethodsAndOrder(t *testing.T) {
	var order []string
	mw := func(name string) middleware.HttpMiddleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name+"+")
				next.ServeHTTP(w, r)
				order = append(order, name+"-")
			})
		}
	}
	s := NewServer(SetHTTPMiddlewares(mw("global")))
	g := s.NewRouteGroup("/admin", mw("outer"))
	child := g.Group("/v1", mw("inner"))
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { order = append(order, "handler"); w.WriteHeader(204) })
	child.Handle("/handle", h)
	child.HandleFunc("/func", h)
	child.HandlePrefix("/files/", h)
	child.HandleHeader(h, "X-Select", "yes")
	child.GET("/method/{id}", h, mw("route"))
	g.Use(mw("late"))
	for _, path := range []string{"/admin/v1/handle", "/admin/v1/func", "/admin/v1/files/a", "/admin/v1/header"} {
		order = nil
		req := httptest.NewRequest("GET", path, nil)
		if strings.HasSuffix(path, "header") {
			req.Header.Set("X-Select", "yes")
		}
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)
		want := []string{"global+", "outer+", "late+", "inner+", "handler", "inner-", "late-", "outer-", "global-"}
		if w.Code != 204 || !reflect.DeepEqual(order, want) {
			t.Fatalf("%s: status=%d order=%v", path, w.Code, order)
		}
	}
	for _, path := range []string{"/handle", "/func", "/files/a", "/administrator/v1/header"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("X-Select", "yes")
		s.ServeHTTP(w, req)
		if w.Code != 404 {
			t.Errorf("leaked route %s: %d", path, w.Code)
		}
	}
	order = nil
	w := httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/admin/v1/method/1", nil))
	want := []string{"global+", "outer+", "late+", "inner+", "route+", "handler", "route-", "inner-", "late-", "outer-", "global-"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order=%v", order)
	}
	var routes []RouteInfo
	if err := s.GetRouter().WalkRoute(func(r RouteInfo) error { routes = append(routes, r); return nil }); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(routes, []RouteInfo{{Path: "/admin/v1/method/{id}", Method: "GET"}}) {
		t.Fatalf("routes=%v", routes)
	}
	w = httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("POST", "/admin/v1/method/1", nil))
	if w.Code != 405 {
		t.Fatalf("wrong method=%d", w.Code)
	}
	child.GET("/trailing/", h)
	w = httptest.NewRecorder()
	s.ServeHTTP(w, httptest.NewRequest("GET", "/admin/v1/trailing", nil))
	if w.Code != 301 {
		t.Fatalf("trailing slash redirect=%d", w.Code)
	}
}

func TestEncodeURLWithError(t *testing.T) {
	for _, id := range []string{"a/b", "a?x=1#fragment", "100%", "中文"} {
		got, err := EncodeURLWithError("/items/{id}?existing=1", struct {
			ID string `json:"id"`
			Q  string `json:"q"`
		}{id, "a b"}, true)
		if err != nil {
			t.Fatal(err)
		}
		u, err := url.Parse(got)
		if err != nil {
			t.Fatal(err)
		}
		if u.EscapedPath() != "/items/"+url.PathEscape(id) || u.Query().Get("q") != "a b" || u.Query().Get("existing") != "1" || u.Fragment != "" {
			t.Errorf("URL=%s", got)
		}
	}
	for _, in := range []interface{}{nil, struct{}{}, struct {
		ID string `json:"id"`
	}{""}} {
		if _, err := EncodeURLWithError("/items/{id}", in, false); err == nil {
			t.Errorf("missing parameter accepted for %#v", in)
		}
	}
	if _, err := EncodeURLWithError("/{id=**}", struct{}{}, false); err == nil {
		t.Error("complex template silently accepted")
	}
	if _, err := EncodeURLWithError("/items", &structpb.Struct{Fields: map[string]*structpb.Value{"key": structpb.NewStringValue("value")}}, true); err == nil {
		t.Error("encoding error ignored")
	}
}

func TestNoBodyResponseSemantics(t *testing.T) {
	for _, tc := range []struct {
		method  string
		status  int
		body    string
		wantErr bool
	}{{"GET", 204, "", false}, {"GET", 205, "", false}, {"HEAD", 200, "", false}, {"HEAD", 404, "", true}, {"GET", 200, "", true}, {"GET", 200, "broken", true}, {"GET", 500, "", true}} {
		out := struct{ Keep int }{7}
		rsp := &http.Response{StatusCode: tc.status, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(tc.body)), Request: &http.Request{Method: tc.method}}
		err := DefaultResponseDecoder(context.Background(), rsp, &out)
		if (err != nil) != tc.wantErr || out.Keep != 7 {
			t.Errorf("%s %d: err=%v out=%v", tc.method, tc.status, err, out)
		}
	}
}

type failedResponseWriter struct {
	header          http.Header
	writes, headers int
	err             error
}

func (w *failedResponseWriter) Header() http.Header       { return w.header }
func (w *failedResponseWriter) WriteHeader(int)           { w.headers++ }
func (w *failedResponseWriter) Write([]byte) (int, error) { w.writes++; return 0, w.err }
func TestResultDoesNotWriteSecondError(t *testing.T) {
	cause := errors.New("connection failed")
	w := &failedResponseWriter{header: make(http.Header), err: cause}
	req := httptest.NewRequest("GET", "/", nil)
	err := newRouterCoder().ResponseEncoder()(200, w, req, struct{}{})
	if !errors.Is(err, cause) {
		t.Fatalf("lost write error: %v", err)
	}
	w.writes = 0
	w.headers = 0
	NewCtx(w, req).Result(200, struct{}{})
	if w.writes != 1 || w.headers != 1 {
		t.Fatalf("second response: writes=%d headers=%d", w.writes, w.headers)
	}
	recorder := httptest.NewRecorder()
	NewCtx(recorder, req).Result(200, make(chan int))
	if recorder.Code == 200 {
		t.Fatal("encoding failure committed success")
	}
}

func TestEncodedPathParametersOverHTTP(t *testing.T) {
	srv := NewServer(RouterUseEncodedPath())
	srv.GetRouter().GET("/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		ctx := NewCtx(w, r)
		var in struct {
			ID string `json:"id"`
		}
		if err := ctx.BindVars(&in); err != nil {
			t.Error(err)
			w.WriteHeader(400)
			return
		}
		if ctx.Vars().Get("id") != in.ID {
			t.Error("Vars and BindVars differ")
		}
		_ = ctx.JSON(200, in)
	})
	ts := httptest.NewServer(srv)
	defer ts.Close()
	c, err := NewClient(WithEndpoint(ts.URL))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	for _, id := range []string{"a/b", "a?x=1#f", "a%2Fb", "中文"} {
		in := struct {
			ID string `json:"id"`
		}{id}
		path, err := EncodeURLWithError("/items/{id}", in, false)
		if err != nil {
			t.Fatal(err)
		}
		var out struct {
			ID string `json:"id"`
		}
		if err = c.Invoke(context.Background(), "GET", path, nil, &out); err != nil {
			t.Fatal(err)
		}
		if out.ID != id {
			t.Fatalf("round trip changed %q to %q", id, out.ID)
		}
	}
}

func TestGroupMiddlewareRetainsHandlerState(t *testing.T) {
	s := NewServer()
	g := s.NewRouteGroup("/group")
	g.GET("/limited", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	built := 0
	g.Use(func(next http.Handler) http.Handler {
		built++
		calls := 0
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if calls > 1 {
				w.WriteHeader(429)
				return
			}
			next.ServeHTTP(w, r)
		})
	})
	for _, want := range []int{204, 429} {
		w := httptest.NewRecorder()
		s.ServeHTTP(w, httptest.NewRequest("GET", "/group/limited", nil))
		if w.Code != want {
			t.Fatalf("middleware state reset: code=%d want=%d", w.Code, want)
		}
	}
	if built != 1 {
		t.Fatalf("middleware rebuilt %d times", built)
	}
}
