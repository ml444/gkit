package httpx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/ml444/gkit/internal/netx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/internal/lifecycle"
)

var _ http.Handler = (*Server)(nil)

// Server is an HTTP server wrappedCtx.
type Server struct {
	*http.Server
	resourceMu          sync.RWMutex
	bindMu              sync.Mutex
	managed             bool
	legacyUsed          bool
	listener            net.Listener
	tlsConf             *tls.Config
	endpoint            *url.URL
	network             string
	address             string
	timeout             time.Duration
	readTimeout         time.Duration
	readHeaderTimeout   time.Duration
	writeTimeout        time.Duration
	idleTimeout         time.Duration
	maxHeaderBytes      int
	maxRequestBodyBytes int64
	router              IRouter
	routerCfg           *RouterCfg
	httpMiddlewares     []middleware.HttpMiddleware
	middlewares         []middleware.Middleware
	disableTransportCtx bool
}

func NewServer(opts ...ServerOption) *Server {
	srv := &Server{
		network:             "tcp",
		address:             ":5050",
		timeout:             1 * time.Second,
		maxRequestBodyBytes: 4 << 20,
		routerCfg:           NewRouterCfg(),
	}
	for _, o := range opts {
		o(srv)
	}
	srv.router = newRouter("/", srv.routerCfg)
	srv.router.Use(srv.globalMiddleware())
	srv.router.Use(srv.httpMiddlewares...)
	srv.Server = &http.Server{
		Handler:           srv.router,
		TLSConfig:         srv.tlsConf,
		ReadTimeout:       srv.readTimeout,
		ReadHeaderTimeout: srv.readHeaderTimeout,
		WriteTimeout:      srv.writeTimeout,
		IdleTimeout:       srv.idleTimeout,
		MaxHeaderBytes:    srv.maxHeaderBytes,
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
			log.Debugf("[HTTP] [%s]%s", req.Method, req.URL.Path)
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
			ctx = context.WithValue(ctx, routerCoderKey{}, s.routerCfg.Coder)
			if s.routerCfg.UseEncodedPath {
				ctx = context.WithValue(ctx, encodedPathKey{}, true)
			}
			if s.maxRequestBodyBytes > 0 && req.Body != nil {
				req.Body = http.MaxBytesReader(w, req.Body, s.maxRequestBodyBytes)
			}
			if !s.disableTransportCtx {
				pathTemplate := req.URL.Path
				if route := mux.CurrentRoute(req); route != nil {
					// /path/123 -> /path/{id}
					pathTemplate, _ = route.GetPathTemplate()
				}
				tr := &Transport{
					path:         pathTemplate,
					pathTemplate: pathTemplate,
					inMD:         transport.New(req.Header),
					outMD:        transport.MD{},
					req:          req,
				}
				tr.endpoint = s.endpointString()
				req = req.WithContext(transport.ToContext(ctx, tr))
				tw := newTransportResponseWriter(w, tr)
				next.ServeHTTP(tw, req)
				tw.flushTransportHeaders()
				return
			} else {
				req = req.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	}
}

// Endpoint binds if necessary and returns an address copy. After Managed has
// taken ownership, use its Listen and EndpointURL methods instead.
func (s *Server) Endpoint() (*url.URL, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return nil, err
	}
	return s.endpointSnapshot(), nil
}

// Start the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	if err := s.useLegacyLifecycle(); err != nil {
		return err
	}
	if err := s.listenAndEndpoint(); err != nil {
		return err
	}
	err := s.serve(ctx)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) serve(ctx context.Context) error {
	lis := s.listenerSnapshot()
	s.BaseContext = func(net.Listener) context.Context {
		return ctx
	}
	log.Infof("[HTTP] server listening on: %s \n", lis.Addr().String())
	if s.tlsConf != nil {
		return s.ServeTLS(lis, "", "")
	}
	return s.Serve(lis)
}

// Stop the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.useLegacyLifecycle(); err != nil {
		return err
	}
	log.Info("[HTTP] server stopping")
	err := s.Shutdown(ctx)
	// Shutdown closes listeners the server is actively serving on; close any
	// listener created early (e.g. via Endpoint()) but never served, so the
	// port is always released.
	_ = lifecycle.CloseListener(s.listenerSnapshot())
	return err
}

func (s *Server) listenAndEndpoint() error {
	return s.bind(context.Background(), false)
}

func (s *Server) bind(ctx context.Context, managed bool) error {
	s.bindMu.Lock()
	defer s.bindMu.Unlock()
	s.resourceMu.RLock()
	owned, lis, endpoint := s.managed, s.listener, s.endpoint
	s.resourceMu.RUnlock()
	if owned != managed {
		return transport.ErrLifecycleOwned
	}
	created := lis == nil
	if created {
		var cfg net.ListenConfig
		var err error
		lis, err = cfg.Listen(ctx, s.network, s.address)
		if err != nil {
			return err
		}
	}
	if endpoint == nil {
		addr, err := netx.ExtractEndpoint(s.address, lis)
		if err != nil {
			if created {
				_ = lis.Close()
			}
			return err
		}
		scheme := "http"
		if s.tlsConf != nil {
			scheme = "https"
		}
		endpoint = &url.URL{Scheme: scheme, Host: addr}
	}
	if err := ctx.Err(); err != nil {
		if created {
			_ = lis.Close()
		}
		return err
	}
	s.resourceMu.Lock()
	s.listener, s.endpoint = lis, endpoint
	s.resourceMu.Unlock()
	return nil
}

func (s *Server) listenerSnapshot() net.Listener {
	s.resourceMu.RLock()
	defer s.resourceMu.RUnlock()
	return s.listener
}

func (s *Server) endpointSnapshot() *url.URL {
	s.resourceMu.RLock()
	defer s.resourceMu.RUnlock()
	return lifecycle.CloneURL(s.endpoint)
}

func (s *Server) endpointString() string {
	s.resourceMu.RLock()
	defer s.resourceMu.RUnlock()
	if s.endpoint == nil {
		return ""
	}
	return s.endpoint.String()
}

func (s *Server) useLegacyLifecycle() error {
	s.resourceMu.Lock()
	defer s.resourceMu.Unlock()
	if s.managed {
		return transport.ErrLifecycleOwned
	}
	s.legacyUsed = true
	return nil
}
