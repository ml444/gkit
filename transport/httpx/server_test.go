package httpx

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/internal/netx"
	"github.com/ml444/gkit/transport"
)

var testHandle = func(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("testHandle:", r.RequestURI)
	_ = json.NewEncoder(w).Encode(testData{Path: r.RequestURI})
}

type testKey struct{}

type testData struct {
	Path string `json:"path"`
}

// handleFuncWrapper is a wrapper for http.HandlerFunc to implement http.Handler
type handleFuncWrapper struct {
	fn http.HandlerFunc
}

func (x *handleFuncWrapper) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	x.fn.ServeHTTP(writer, request)
}

func newHandleFuncWrapper(fn http.HandlerFunc) http.Handler {
	fmt.Println("======> test <======")
	return &handleFuncWrapper{fn: fn}
}

func TestServeHTTP(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	svr := NewServer(Listener(ln))
	router := svr.GetRouter()
	router.HandleFunc("/index", testHandle)
	router.Group("/errors").GET("/cause", func(w http.ResponseWriter, r *http.Request) {
		decoder := NewCtx(w, r)
		err := errorx.BadRequest("zzz").
			WithMetadata(map[string]string{"foo": "bar"}).
			WithCause(errors.New("error cause"))
		decoder.ReturnError(err)
	})
	if err = router.WalkRoute(func(r RouteInfo) error {
		t.Logf("WalkRoute: %+v", r)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if e, err := svr.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	}
	srv := http.Server{Handler: svr}
	go func() {
		if err := srv.Serve(ln); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Log(err)
	}
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	srv := NewServer()
	router := srv.GetRouter()
	router.Handle("/index", newHandleFuncWrapper(testHandle))
	router.HandleFunc("/index/{id:[0-9]+}", testHandle)
	router.HandlePrefix("/test/prefix", newHandleFuncWrapper(testHandle))
	router.HandleHeader(testHandle, "content-type", "application/grpc-web+json")
	router.Group("/errors").GET("/cause", func(w http.ResponseWriter, r *http.Request) {
		decoder := NewCtx(w, r)
		err := errorx.BadRequest("zzz").
			WithMetadata(map[string]string{"foo": "bar"}).
			WithCause(errors.New("error cause"))
		decoder.ReturnError(err)
	})

	if e, err := srv.Endpoint(); err != nil || e == nil || strings.HasSuffix(e.Host, ":0") {
		t.Fatal(e, err)
	}

	go func() {
		if err := srv.Start(ctx); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	testHeader(t, srv)
	testClient(t, srv)
	testAccept(t, srv)
	time.Sleep(time.Second)
	if srv.Stop(ctx) != nil {
		t.Errorf("expected nil got %v", srv.Stop(ctx))
	}
}

func TestServerFlushesTransportHeadersBeforeWrite(t *testing.T) {
	srv := NewServer()
	srv.GetRouter().GET("/headers", func(w http.ResponseWriter, r *http.Request) {
		tr, ok := transport.FromContext(r.Context())
		if !ok {
			t.Fatal("missing transport context")
		}
		tr.Out().Set("X-Trace-ID", "trace-1")
		_, _ = w.Write([]byte("ok"))
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/headers", nil)
	srv.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Trace-ID"); got != "trace-1" {
		t.Fatalf("X-Trace-ID = %q, want trace-1", got)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Fatalf("body = %q, want ok", got)
	}
}

func TestServerFlushesTransportHeadersWhenHandlerDoesNotWrite(t *testing.T) {
	srv := NewServer()
	srv.GetRouter().GET("/headers", func(w http.ResponseWriter, r *http.Request) {
		tr, ok := transport.FromContext(r.Context())
		if !ok {
			t.Fatal("missing transport context")
		}
		tr.Out().Set("X-Trace-ID", "trace-2")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/headers", nil)
	srv.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Trace-ID"); got != "trace-2" {
		t.Fatalf("X-Trace-ID = %q, want trace-2", got)
	}
}

func testAccept(t *testing.T, srv *Server) {
	tests := []struct {
		method      string
		path        string
		contentType string
	}{
		{http.MethodGet, "/errors/cause", "application/json"},
		{http.MethodGet, "/errors/cause", "application/proto"},
	}
	e, err := srv.Endpoint()
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	client, err := NewClient(WithEndpoint(e.Host))
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	for _, test := range tests {
		req, err := http.NewRequest(test.method, e.String()+test.path, nil)
		if err != nil {
			t.Errorf("expected nil got %v", err)
		}
		req.Header.Set("Content-Type", test.contentType)
		resp, req, err := client.Do(req)
		if err != nil {
			t.Errorf("expected nil got %v", err)
			continue
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 got %v", resp.StatusCode)
		}
		resp.Body.Close()
	}
}

func testHeader(t *testing.T, srv *Server) {
	e, err := srv.Endpoint()
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	client, err := NewClient(WithEndpoint(e.Host))
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	reqURL := fmt.Sprintf(e.String() + "/index")
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		t.Errorf("expected nil got %v", err)
	}
	req.Header.Set("content-type", "application/grpc-web+json")
	resp, req, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		t.Errorf("expected nil got %v", err)
	}
	resp.Body.Close()
}

func testClient(t *testing.T, srv *Server) {
	tests := []struct {
		method string
		path   string
		code   int
	}{
		{http.MethodGet, "/index", http.StatusOK},
		{http.MethodPut, "/index", http.StatusOK},
		{http.MethodPost, "/index", http.StatusOK},
		{http.MethodPatch, "/index", http.StatusOK},
		{http.MethodDelete, "/index", http.StatusOK},

		{http.MethodGet, "/index/1", http.StatusOK},
		{http.MethodPut, "/index/1", http.StatusOK},
		{http.MethodPost, "/index/1", http.StatusOK},
		{http.MethodPatch, "/index/1", http.StatusOK},
		{http.MethodDelete, "/index/1", http.StatusOK},

		{http.MethodGet, "/index/notfound", http.StatusNotFound},
		{http.MethodGet, "/errors/cause", http.StatusBadRequest},
		{http.MethodGet, "/test/prefix/123111", http.StatusOK},
	}
	e, err := srv.Endpoint()
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewClient(
		WithEndpoint(e.Host))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, test := range tests {
		var res testData
		reqURL := fmt.Sprintf(e.String() + test.path)
		req, err := http.NewRequest(test.method, reqURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		//req.Header.Set("content-type", "application/json")
		resp, req, err := client.Do(req)
		if err != nil {
			t.Fatalf("want %v, but got %v", test, err)
		}
		if resp.StatusCode != test.code {
			_ = resp.Body.Close()
			t.Fatalf("want %v, but got %v", test, err)
		}
		if test.code >= http.StatusBadRequest {
			_ = resp.Body.Close()
			continue
		}
		content, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			t.Fatalf("read resp error %v", err)
		}
		err = json.Unmarshal(content, &res)
		if err != nil {
			t.Log(string(content))
			t.Fatalf("unmarshal resp error %v, test data: %v", err, test)
		}
		if res.Path != test.path {
			t.Errorf("expected %s got %s", test.path, res.Path)
		}
	}
	for _, test := range tests {
		var res testData
		err := client.Invoke(context.Background(), test.method, test.path, nil, &res)
		if errorx.Status(err) != test.code {
			t.Fatalf("want %v, but got %v", test, err)
		}
		if err != nil {
			continue
		}
		if res.Path != test.path {
			t.Errorf("expected %s got %s", test.path, res.Path)
		}
	}
}

func BenchmarkServer(b *testing.B) {
	fn := func(w http.ResponseWriter, r *http.Request) {
		data := &testData{Path: r.RequestURI}
		_ = json.NewEncoder(w).Encode(data)
		if r.Context().Value(testKey{}) != "test" {
			w.WriteHeader(500)
		}
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, testKey{}, "test")
	srv := NewServer()
	srv.GetRouter().HandleFunc("/index", fn)
	go func() {
		if err := srv.Start(ctx); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	port, ok := netx.Port(srv.listener)
	if !ok {
		b.Errorf("expected port got %v", srv.listener)
	}
	client, err := NewClient(WithEndpoint(fmt.Sprintf("127.0.0.1:%d", port)))
	if err != nil {
		b.Errorf("expected nil got %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var res testData
		err := client.Invoke(context.Background(), http.MethodPost, "/index", nil, &res)
		if err != nil {
			b.Errorf("expected nil got %v", err)
		}
	}
	_ = srv.Stop(ctx)
}

func TestNetwork(t *testing.T) {
	o := &Server{}
	v := "abc"
	Network(v)(o)
	if !reflect.DeepEqual(v, o.network) {
		t.Errorf("expected %v got %v", v, o.network)
	}
}

func TestServerEndpointAndStartErrors(t *testing.T) {
	srv := NewServer(Network("bad-network"))
	if _, err := srv.Endpoint(); err == nil {
		t.Fatal("expected endpoint listen error")
	}
	if err := srv.Start(context.Background()); err == nil {
		t.Fatal("expected start listen error")
	}
}

func TestAddress(t *testing.T) {
	o := &Server{}
	v := "abc"
	Address(v)(o)
	if !reflect.DeepEqual(v, o.address) {
		t.Errorf("expected %v got %v", v, o.address)
	}
}

func TestTimeout(t *testing.T) {
	o := &Server{}
	v := time.Duration(123)
	Timeout(v)(o)
	if !reflect.DeepEqual(v, o.timeout) {
		t.Errorf("expected %v got %v", v, o.timeout)
	}
}

func TestTLSConfig(t *testing.T) {
	o := &Server{}
	v := &tls.Config{}
	TLSConfig(v)(o)
	if !reflect.DeepEqual(v, o.tlsConf) {
		t.Errorf("expected %v got %v", v, o.tlsConf)
	}
}

func TestListener(t *testing.T) {
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{}
	Listener(lis)(s)
	if !reflect.DeepEqual(s.listener, lis) {
		t.Errorf("expected %v got %v", lis, s.listener)
	}
	if e, err := s.Endpoint(); err != nil || e == nil {
		t.Errorf("expected not empty")
	}
}
