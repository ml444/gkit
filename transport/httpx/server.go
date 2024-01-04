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

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/httpx/host"
)

var (
	_ transport.Server = (*Server)(nil)
	//_ transport.Endpointer = (*Server)(nil)
	_ http.Handler = (*Server)(nil)
)

// Server is an HTTP server wrapper.
type Server struct {
	*http.Server
	lis         net.Listener
	tlsConf     *tls.Config
	endpoint    *url.URL
	err         error
	network     string
	address     string
	timeout     time.Duration
	router      *mux.Router
	middlewares []middleware.Middleware
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network: "tcp",
		address: ":5050",
		timeout: 1 * time.Second,
		router:  NewMuxRouter(),
	}
	for _, o := range opts {
		o(srv)
	}
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

// WalkRoute walks the router and all its sub-routers, calling walkFn for each route in the tree.
func (s *Server) WalkRoute(fn WalkRouteFunc) error {
	return s.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		methods, err := route.GetMethods()
		if err != nil {
			return nil // ignore no methods
		}
		path, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		for _, method := range methods {
			if err := fn(RouteInfo{Method: method, Path: path}); err != nil {
				return err
			}
		}
		return nil
	})
}

// WalkHandle walks the router and all its sub-routers, calling walkFn for each route in the tree.
func (s *Server) WalkHandle(handle func(method, path string, handler http.HandlerFunc)) error {
	return s.WalkRoute(func(r RouteInfo) error {
		handle(r.Method, r.Path, s.ServeHTTP)
		return nil
	})
}

// Route registers an HTTP router.
func (s *Server) Route(prefix string, httpMiddlewares ...middleware.HttpMiddleware) *Router {
	return newRouter(prefix, s.router, DefaultErrorEncoder, httpMiddlewares...)
}

// Handle registers a new route with a matcher for the URL path.
func (s *Server) Handle(method, path string, h http.Handler) {
	s.router.Handle(path, h).Methods(method)
}

// HandlePrefix registers a new route with a matcher for the URL path prefix.
func (s *Server) HandlePrefix(method, prefix string, h http.Handler) {
	s.router.PathPrefix(prefix).Handler(h).Methods(method)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (s *Server) HandleFunc(method, path string, h http.HandlerFunc) {
	s.router.HandleFunc(path, h).Methods(method)
}

// HandleHeader registers a new route with a matcher for the header.
func (s *Server) HandleHeader(method string, h http.HandlerFunc, headerPairs ...string) {
	s.router.Headers(headerPairs...).Handler(h).Methods(method)
}

// ServeHTTP should write reply headers and data to the ResponseWriter and then return.
func (s *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	s.Handler.ServeHTTP(res, req)
}

func (s *Server) globalMiddleware() mux.MiddlewareFunc {
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

			pathTemplate := req.URL.Path
			if route := mux.CurrentRoute(req); route != nil {
				// /path/123 -> /path/{id}
				pathTemplate, _ = route.GetPathTemplate()
			}
			log.Debugf("%s %s\n", req.Method, pathTemplate)
			tr := &transport.Transport{
				Operation: pathTemplate,
				//pathTemplate: pathTemplate,
				InHeader:  header.New(req.Header),
				OutHeader: header.New(w.Header()),
				//request:      req,
			}
			if s.endpoint != nil {
				tr.Endpoint = s.endpoint.String()
			}
			tr.Request = req.WithContext(transport.ToContext(ctx, tr))
			next.ServeHTTP(w, tr.Request)
		})
	}
}

// Endpoint return a real address to registry endpoint.
// examples:
//
//	https://127.0.0.1:8000
//	Legacy: http://127.0.0.1:8000?isSecure=false
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
	log.Infof("[HTTP] server listening on: %s", s.lis.Addr().String())
	var err error
	if s.tlsConf != nil {
		err = s.ServeTLS(s.lis, "", "")
	} else {
		err = s.Serve(s.lis)
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
	if s.lis == nil {
		lis, err := net.Listen(s.network, s.address)
		if err != nil {
			s.err = err
			return err
		}
		s.lis = lis
	}
	if s.endpoint == nil {
		addr, err := host.Extract(s.address, s.lis)
		if err != nil {
			s.err = err
			return err
		}
		s.endpoint = &url.URL{Scheme: Scheme("http", s.tlsConf != nil), Host: addr}
	}
	return s.err
}

func Scheme(scheme string, isSecure bool) string {
	if isSecure {
		return scheme + "s"
	}
	return scheme
}
