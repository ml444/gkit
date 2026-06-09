package httpx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/middleware"
)

func TestServerOptions(t *testing.T) {
	s := &Server{routerCfg: NewRouterCfg()}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	Network("tcp4")(s)
	Address("127.0.0.1:0")(s)
	Endpoint(nil)(s)
	Timeout(time.Second)(s)
	MaxRequestBodySize(123)(s)
	TLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})(s)
	Listener(lis)(s)
	DisableTransportCtx()(s)
	SetMiddlewares(func(middleware.ServiceHandler) middleware.ServiceHandler { return nil })(s)
	Middleware()(s)
	SetHTTPMiddlewares(func(next http.Handler) http.Handler { return next })(s)
	RouterPathPrefix("/api")(s)
	RouterStrictSlash(false)(s)
	RouterSkipClean(true)(s)
	RouterUseEncodedPath()(s)
	RouterNotFoundHandler(http.NotFoundHandler())(s)
	RouterMethodNotAllowedHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))(s)
	RouterCoder(newRouterCoder())(s)

	if s.network != "tcp4" || s.address != "127.0.0.1:0" || s.timeout != time.Second ||
		s.maxRequestBodyBytes != 123 || s.tlsConf == nil || s.listener != lis ||
		!s.disableTransportCtx || len(s.middlewares) != 1 || len(s.httpMiddlewares) != 1 ||
		s.routerCfg.RootPrefix != "/api" || s.routerCfg.StrictSlash || !s.routerCfg.SkipClean ||
		!s.routerCfg.UseEncodedPath || s.routerCfg.Coder == nil {
		t.Fatalf("server options not applied: %#v %#v", s, s.routerCfg)
	}
}

func TestClientOptionsAndParseTarget(t *testing.T) {
	dc := discovery.NewDiscoveryClient(discovery.NewDefaultRegistry())
	c := &Client{}
	rt := roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("stop")
	})
	WithTransport(rt)(c)
	WithTimeout(time.Second)(c)
	WithUserAgent("ua")(c)
	WithMiddlewares(func(next middleware.ServiceHandler) middleware.ServiceHandler { return next })(c)
	WithEndpoint("example.com")(c)
	WithDiscovery(dc, "svc")(c)
	WithRequestEncoder(func(context.Context, string, interface{}) ([]byte, error) { return []byte("x"), nil })(c)
	WithResponseDecoder(func(context.Context, *http.Response, interface{}) error { return nil })(c)
	WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})(c)

	if c.transport == nil || c.timeout != time.Second || c.userAgent != "ua" ||
		len(c.middleware) != 1 || c.endpoint != "example.com" || c.discovery != dc ||
		c.service != "svc" || c.encoder == nil || c.decoder == nil || c.tlsConf == nil {
		t.Fatalf("client options not applied: %#v", c)
	}

	tests := []struct {
		endpoint string
		insecure bool
		scheme   string
		auth     string
		svc      string
	}{
		{"example.com", true, "http", "example.com", ""},
		{"https://example.com", false, "https", "example.com", ""},
		{"discovery:///svc", true, "http", "discovery", "svc"},
	}
	for _, tt := range tests {
		target, err := parseTarget(tt.endpoint, tt.insecure)
		if err != nil {
			t.Fatalf("parseTarget(%q): %v", tt.endpoint, err)
		}
		if target.Scheme != tt.scheme || target.Authority != tt.auth || target.DiscoveryService != tt.svc {
			t.Fatalf("target = %#v", target)
		}
	}
}

func TestCallOptionsAndEncodeURL(t *testing.T) {
	ci := defaultCallInfo("/v1/users/{id}")
	SetRequestContentType("application/proto")(&ci)
	RequestHeader(http.Header{"X-One": {"1"}})(&ci)
	AddRequestHeader("X-One", "2")(&ci)
	Operation("op")(&ci)
	PathTemplate("/tpl/{id}")(&ci)
	OnResponse(func(*http.Response) error { return nil })(&ci)
	if ci.reqHeader.Get("Content-Type") != "" || ci.reqHeader.Values("X-One")[1] != "2" ||
		ci.operation != "op" || ci.pathTemplate != "/tpl/{id}" || ci.onResponse == nil {
		t.Fatalf("call info = %#v", ci)
	}

	type req struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if got := EncodeURL("/v1/users/{id}", req{ID: "42", Name: "neo"}, true); got != "/v1/users/42?name=neo" {
		t.Fatalf("encoded url = %q", got)
	}
	if got := EncodeURL("/v1/users/{id}", (*req)(nil), true); got != "/v1/users/{id}" {
		t.Fatalf("nil encoded url = %q", got)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
