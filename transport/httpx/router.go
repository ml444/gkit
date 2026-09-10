package httpx

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"

	"github.com/ml444/gkit/middleware"
)

type IRouter interface {
	IHttpRouter
	IRouteMethod
	WalkRoute(fn WalkRouteFunc) error
}

type IHttpRouter interface {
	ServeHTTP(res http.ResponseWriter, req *http.Request)
	Use(mws ...middleware.HttpMiddleware)
	Group(prefix string, middlewares ...middleware.HttpMiddleware) *Router
	Handle(path string, h http.Handler)
	HandlePrefix(prefix string, h http.Handler)
	HandleFunc(path string, h http.HandlerFunc)
	HandleHeader(h http.HandlerFunc, headerPairs ...string)
}

type IRouteMethod interface {
	GET(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
	HEAD(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
	POST(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
	PUT(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
	PATCH(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
	DELETE(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
	CONNECT(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
	OPTIONS(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
	TRACE(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware)
}

type IRouterSetting interface {
	UseEncodedPath() bool
	SetUseEncodedPath(bool)
	StrictSlash() bool
	SetStrictSlash(bool)
	SkipClean() bool
	SetSkipClean(bool)
	Prefix() string
	SetPrefix(string)
	NotFoundHandler() http.Handler
	SetNotFoundHandler(handler http.Handler)
	NotAllowedHandler() http.Handler
	SetNotAllowedHandler(handler http.Handler)
}

// HandleFunc defines a function to serve HTTP requests.
//type HandleFunc func(Context) error

//type HandleFunc func(ctx context.Context, req interface{}) (interface{}, error)

var _ = IRouter(&Router{})

func NewRouterCfg() *RouterCfg {
	return &RouterCfg{
		UseEncodedPath: false,
		StrictSlash:    true,
		SkipClean:      false,
		RootPrefix:     "",
		NotFoundHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
		MethodNotAllowedHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}),
		Coder: newRouterCoder(),
	}
}

// Router is an HTTP coder.
type Router struct {
	*RouterCfg
	prefix      string
	router      *mux.Router
	middlewares []middleware.HttpMiddleware
	parent      *Router
}

type RouterCfg struct {
	// If true, "/path/foo%2Fbar/to" will match the path "/path/{var}/to"
	UseEncodedPath bool

	// If true, when the path pattern is "/path/", accessing "/path" will
	// redirect to the former and vice versa.
	StrictSlash bool

	// If true, when the path pattern is "/path//to", accessing "/path//to"
	// will not redirect
	SkipClean bool

	// Configurable Handler to be used when no route matches.
	NotFoundHandler http.Handler

	// Configurable Handler to be used when the request method does not match the route.
	MethodNotAllowedHandler http.Handler

	RootPrefix string

	Coder IRouterCoder
}

// WalkRouteFunc is the type of the function called for each route visited by Walk.
type WalkRouteFunc func(RouteInfo) error

// RouteInfo is an HTTP route info.
type RouteInfo struct {
	Path   string
	Method string
}

func newMuxRouter(cfg *RouterCfg) *mux.Router {
	r := mux.NewRouter()
	if cfg.RootPrefix != "" {
		r = r.PathPrefix(cfg.RootPrefix).Subrouter()
	}
	if cfg.UseEncodedPath {
		r.UseEncodedPath()
	}
	r.SkipClean(cfg.SkipClean)
	r.StrictSlash(cfg.StrictSlash)
	r.NotFoundHandler = cfg.NotFoundHandler
	r.MethodNotAllowedHandler = cfg.MethodNotAllowedHandler
	return r
}
func newRouter(prefix string, cfg *RouterCfg, middlewares ...middleware.HttpMiddleware) *Router {
	if cfg.Coder == nil {
		cfg.Coder = newRouterCoder()
	}
	r := &Router{
		prefix:      prefix,
		router:      newMuxRouter(cfg),
		middlewares: middlewares,
		RouterCfg:   cfg,
	}
	return r
}

// Use applies middleware to this group and its descendants. Configure before serving.
func (r *Router) Use(mws ...middleware.HttpMiddleware) {
	if r.parent != nil {
		r.middlewares = append(r.middlewares, mws...)
		return
	}
	var mwList []mux.MiddlewareFunc
	for _, mw := range mws {
		mwList = append(mwList, mux.MiddlewareFunc(mw))
	}
	r.router.Use(mwList...)
}

func (r *Router) Handle(path string, h http.Handler) {
	r.router.Handle(r.routePath(path), r.wrap(h))
}

func (r *Router) HandlePrefix(prefix string, h http.Handler) {
	if r.parent != nil && prefix == "" {
		base := strings.TrimSuffix(r.prefix, "/")
		next := r.wrap(h)
		r.router.Path(base).Handler(next)
		r.router.PathPrefix(base + "/").Handler(next)
		return
	}
	r.router.PathPrefix(r.routePath(prefix)).Handler(r.wrap(h))
}

func (r *Router) HandleFunc(path string, h http.HandlerFunc) {
	r.Handle(path, h)
}

func (r *Router) HandleHeader(h http.HandlerFunc, headerPairs ...string) {
	next := r.wrap(h)
	if r.parent == nil {
		r.router.Headers(headerPairs...).Handler(next)
		return
	}
	prefix := strings.TrimSuffix(r.prefix, "/")
	r.router.Path(prefix).Headers(headerPairs...).Handler(next)
	r.router.PathPrefix(prefix + "/").Headers(headerPairs...).Handler(next)
}
func (r *Router) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(res, req)
}

func (r *Router) WalkRoute(fn WalkRouteFunc) error {
	return r.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		methods, err := route.GetMethods()
		if err != nil {
			return nil // ignore no methods
		}
		pathTpml, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		for _, method := range methods {
			if err := fn(RouteInfo{Method: method, Path: pathTpml}); err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *Router) routePath(relativePath string) string {
	joined := JoinPath(r.prefix, relativePath)
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}

func (r *Router) Group(prefix string, middlewares ...middleware.HttpMiddleware) *Router {
	return &Router{
		prefix: r.routePath(prefix), router: r.router,
		middlewares: append([]middleware.HttpMiddleware(nil), middlewares...),
		RouterCfg:   r.RouterCfg, parent: r,
	}
}

// wrap builds the group chain once on first use so middleware retains its state.
// Use also applies to routes registered earlier, before serving starts.
// Like mux.Router, route configuration must be complete before serving requests.
func (r *Router) wrap(h http.Handler) http.Handler {
	var once sync.Once
	var next http.Handler
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		once.Do(func() {
			next = h
			for group := r; group != nil; group = group.parent {
				next = middleware.HTTPChain(group.middlewares...)(next)
			}
		})
		next.ServeHTTP(w, req)
	})
}

func (r *Router) handle(method, relativePath string, h http.HandlerFunc, middlewares ...middleware.HttpMiddleware) {
	next := middleware.HTTPChain(middlewares...)(h)
	r.router.Handle(r.routePath(relativePath), r.wrap(next)).Methods(method)
}

func (r *Router) GET(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodGet, path, h, m...)
}

func (r *Router) HEAD(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodHead, path, h, m...)
}

func (r *Router) POST(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodPost, path, h, m...)
}

func (r *Router) PUT(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodPut, path, h, m...)
}

func (r *Router) PATCH(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodPatch, path, h, m...)
}

func (r *Router) DELETE(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodDelete, path, h, m...)
}

func (r *Router) CONNECT(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodConnect, path, h, m...)
}

func (r *Router) OPTIONS(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodOptions, path, h, m...)
}

func (r *Router) TRACE(path string, h http.HandlerFunc, m ...middleware.HttpMiddleware) {
	r.handle(http.MethodTrace, path, h, m...)
}
