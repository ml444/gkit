package httpx

import (
	"net/http"
	"path"
	"sync"

	"github.com/ml444/gkit/middleware"
)

// WalkRouteFunc is the type of the function called for each route visited by Walk.
type WalkRouteFunc func(RouteInfo) error

// RouteInfo is an HTTP route info.
type RouteInfo struct {
	Path   string
	Method string
}

// HandlerFunc defines a function to serve HTTP requests.
type HandlerFunc func(Context) error

// Router is an HTTP router.
type Router struct {
	prefix      string
	pool        sync.Pool
	srv         *Server
	middlewares []middleware.HttpMiddleware
}

func newRouter(prefix string, srv *Server, middlewares ...middleware.HttpMiddleware) *Router {
	r := &Router{
		prefix:      prefix,
		srv:         srv,
		middlewares: middlewares,
	}
	r.pool.New = func() interface{} {
		return &wrapper{router: r}
	}
	return r
}

// Group returns a new router group.
func (r *Router) Group(prefix string, middlewares ...middleware.HttpMiddleware) *Router {
	var newFilters []middleware.HttpMiddleware
	newFilters = append(newFilters, r.middlewares...)
	newFilters = append(newFilters, middlewares...)
	return newRouter(path.Join(r.prefix, prefix), r.srv, newFilters...)
}

// Handle registers a new route with a matcher for the URL path and method.
func (r *Router) Handle(method, relativePath string, h HandlerFunc, middlewares ...middleware.HttpMiddleware) {
	next := http.Handler(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		ctx := r.pool.Get().(Context)
		ctx.Reset(res, req)
		if err := h(ctx); err != nil {
			DefaultErrorEncoder(res, req, err) // DefaultErrorEncoder
		}
		ctx.Reset(nil, nil)
		r.pool.Put(ctx)
	}))
	next = middleware.Chain(middlewares...)(next)
	next = middleware.Chain(r.middlewares...)(next)
	r.srv.router.Handle(path.Join(r.prefix, relativePath), next).Methods(method)
}

// GET registers a new GET route for a path with matching handler in the router.
func (r *Router) GET(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodGet, path, h, m...)
}

// HEAD registers a new HEAD route for a path with matching handler in the router.
func (r *Router) HEAD(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodHead, path, h, m...)
}

// POST registers a new POST route for a path with matching handler in the router.
func (r *Router) POST(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodPost, path, h, m...)
}

// PUT registers a new PUT route for a path with matching handler in the router.
func (r *Router) PUT(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodPut, path, h, m...)
}

// PATCH registers a new PATCH route for a path with matching handler in the router.
func (r *Router) PATCH(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodPatch, path, h, m...)
}

// DELETE registers a new DELETE route for a path with matching handler in the router.
func (r *Router) DELETE(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodDelete, path, h, m...)
}

// CONNECT registers a new CONNECT route for a path with matching handler in the router.
func (r *Router) CONNECT(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodConnect, path, h, m...)
}

// OPTIONS registers a new OPTIONS route for a path with matching handler in the router.
func (r *Router) OPTIONS(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodOptions, path, h, m...)
}

// TRACE registers a new TRACE route for a path with matching handler in the router.
func (r *Router) TRACE(path string, h HandlerFunc, m ...middleware.HttpMiddleware) {
	r.Handle(http.MethodTrace, path, h, m...)
}
