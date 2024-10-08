package httpx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"

	"github.com/ml444/gkit/internal/netx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

var (
	_ http.Handler = (*Server)(nil)
)

// Server is an HTTP server wrappedCtx.
type Server struct {
	*http.Server
	listener            net.Listener
	tlsConf             *tls.Config
	endpoint            *url.URL
	network             string
	address             string
	timeout             time.Duration
	router              IRouter
	routerCfg           *RouterCfg
	middlewares         []middleware.Middleware
	disableTransportCtx bool
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:   "tcp",
		address:   ":5050",
		timeout:   1 * time.Second,
		routerCfg: NewRouterCfg(),
	}
	for _, o := range opts {
		o(srv)
	}
	srv.router = newRouter("/", srv.routerCfg)
	srv.router.Use(srv.globalMiddleware())
	srv.Server = &http.Server{
		Handler:   srv.router,
		TLSConfig: srv.tlsConf,
	}
	return srv
}

func (s *Server) Middlewares() []middleware.Middleware {
	return s.middlewares
}

func (s *Server) SetMiddlewares(mws ...middleware.Middleware) {
	s.middlewares = append(s.middlewares, mws...)
}

func (s *Server) GetRouter() IRouter {
	return s.router
}

func (s *Server) NewRouteGroup(prefix string, httpMiddlewares ...middleware.HttpMiddleware) *Router {
	return s.router.Group(prefix, httpMiddlewares...)
}

// ServeHTTP should write reply headers and data to the ResponseWriter and then return.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.Handler.ServeHTTP(res, req)
}

func (s *Server) globalMiddleware() middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			var (
				ctx    context.Context
				cancel context.CancelFunc
			)
			if s.timeout > 0 {
				ctx, cancel = context.WithTimeout(req.Context(), s.timeout)
			} else {
				ctx, cancel = context.WithCancel(req.Context())
			}
			defer cancel()
			if !s.disableTransportCtx {
				pathTemplate := req.URL.Path
				if route := mux.CurrentRoute(req); route != nil {
					// /path/123 -> /path/{id}
					pathTemplate, _ = route.GetPathTemplate()
				}
				tr := &Transport{
					path: pathTemplate,
					inMD: transport.MD(req.Header),
					req:  req,
				}
				if s.endpoint != nil {
					tr.endpoint = s.endpoint.String()
				}
				req = req.WithContext(transport.ToContext(ctx, tr))
				defer func() {
					tr.outMD = transport.MD(w.Header())
				}()
			}
			next.ServeHTTP(w, req)
		})
	}
}

func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, err
	}
	return s.endpoint, nil
}

// Start the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if err := s.listenAndEndpoint(); err != nil {
		return err
	}
	s.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	log.Infof("[HTTP] server listening on: %s \n", s.listener.Addr().String())
	var err error
	if s.tlsConf != nil {
		err = s.ServeTLS(s.listener, "", "")
	} else {
		err = s.Serve(s.listener)
	}
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	log.Info("[HTTP] server stopping")
	return s.Shutdown(ctx)
}

func (s *Server) listenAndEndpoint() error {
	if s.listener == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			return err
		}
		s.listener = lis
	}
	if s.endpoint == nil {
		addr, err := netx.ExtractEndpoint(s.address, s.listener)
		if err != nil {
			return err
		}
		scheme := "http"
		if s.tlsConf != nil {
			scheme = "https"
		}
		s.endpoint = &url.URL{Scheme: scheme, Host: addr}
	}
	return nil
}
