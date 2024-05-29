package httpx

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ml444/gkit/middleware"
)

// ServerOption is an HTTP server option.
type ServerOption func(*Server)

// Network with server network.
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Endpoint with server endpoint address.
func Endpoint(endpoint *url.URL) ServerOption {
	return func(s *Server) {
		s.endpoint = endpoint
	}
}

// Timeout with server timeout.
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// TLSConfig with TLS config.
func TLSConfig(c *tls.Config) ServerOption {
	return func(o *Server) {
		o.tlsConf = c
	}
}

// Listener with server listener.
func Listener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.listener = lis
	}
}

// DisableTransportCtx cancel default Transport context handling
func DisableTransportCtx() ServerOption {
	return func(s *Server) {
		s.disableTransportCtx = true
	}
}

// SetMiddlewares with server middlewares.
func SetMiddlewares(mws ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, mws...)
	}
}

// RouterPathPrefix with mux's PathPrefix, coder will replace by a subrouter that start with RootPrefix.
func RouterPathPrefix(prefix string) ServerOption {
	return func(s *Server) {
		s.routerCfg.RootPrefix = prefix
	}
}

// RouterStrictSlash with mux's StrictSlash
// If true, when the path pattern is "/path/", accessing "/path" will
// redirect to the former and vice versa.
func RouterStrictSlash(strictSlash bool) ServerOption {
	return func(s *Server) {
		s.routerCfg.StrictSlash = strictSlash
	}
}

// RouterSkipClean with mux's SkipClean
func RouterSkipClean(skipClean bool) ServerOption {
	return func(s *Server) {
		s.routerCfg.SkipClean = skipClean
	}
}

// RouterUseEncodedPath with mux's SkipClean
func RouterUseEncodedPath() ServerOption {
	return func(s *Server) {
		s.routerCfg.UseEncodedPath = true
	}
}

// RouterNotFoundHandler Set up `404 NOT FOUND` handler function
func RouterNotFoundHandler(handler http.Handler) ServerOption {
	return func(s *Server) {
		s.routerCfg.NotFoundHandler = handler
	}
}

// RouterMethodNotAllowedHandler Set up `405 MethodNotAllowed` handler function
func RouterMethodNotAllowedHandler(handler http.Handler) ServerOption {
	return func(s *Server) {
		s.routerCfg.MethodNotAllowedHandler = handler
	}
}

// RouterCoder Customize request parameter and body decoders as well as normal response and error response encoders
func RouterCoder(coder IRouterCoder) ServerOption {
	return func(s *Server) {
		s.routerCfg.Coder = coder
	}
}

// ClientOption is HTTP client option.
type ClientOption func(*Client)

// WithTransport with client transport.
func WithTransport(trans http.RoundTripper) ClientOption {
	return func(o *Client) {
		o.transport = trans
	}
}

// WithTimeout with client request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(o *Client) {
		o.timeout = d
	}
}

// WithUserAgent with client user agent.
func WithUserAgent(ua string) ClientOption {
	return func(o *Client) {
		o.userAgent = ua
	}
}

// WithMiddlewares with client middleware.
func WithMiddlewares(m ...middleware.Middleware) ClientOption {
	return func(o *Client) {
		o.middleware = m
	}
}

// WithEndpoint with client addr.
func WithEndpoint(endpoint string) ClientOption {
	return func(o *Client) {
		o.endpoint = endpoint
	}
}

// WithRequestEncoder with client request encoder.
func WithRequestEncoder(encoder EncodeRequestFunc) ClientOption {
	return func(o *Client) {
		o.encoder = encoder
	}
}

// WithResponseDecoder with client response decoder.
func WithResponseDecoder(decoder DecodeResponseFunc) ClientOption {
	return func(o *Client) {
		o.decoder = decoder
	}
}

// WithTLSConfig with tls config.
func WithTLSConfig(c *tls.Config) ClientOption {
	return func(o *Client) {
		o.tlsConf = c
	}
}

type CallOption func(*callInfo)

type callInfo struct {
	reqHeader    http.Header
	operation    string
	pathTemplate string
	onResponse   func(*http.Response) error
}

func defaultCallInfo(path string) callInfo {
	return callInfo{
		reqHeader: http.Header{
			"Content-Type": []string{"application/json"},
		},
		operation:    path,
		pathTemplate: path,
	}
}

// SetRequestContentType with request content type.
func SetRequestContentType(contentType string) CallOption {
	return func(info *callInfo) {
		info.reqHeader.Set("Content-Type", contentType)
	}
}

func RequestHeader(header http.Header) CallOption {
	return func(ci *callInfo) {
		ci.reqHeader = header
	}
}

func AddRequestHeader(key, value string) CallOption {
	return func(ci *callInfo) {
		ci.reqHeader.Add(key, value)
	}
}

func Operation(operation string) CallOption {
	return func(info *callInfo) {
		info.operation = operation
	}
}

// PathTemplate is http path template
func PathTemplate(pattern string) CallOption {
	return func(info *callInfo) {
		info.pathTemplate = pattern
	}
}

func OnResponse(onResponse func(*http.Response) error) CallOption {
	return func(info *callInfo) {
		info.onResponse = onResponse
	}
}
