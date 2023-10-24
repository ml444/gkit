package middleware

import (
	"context"
	"net/http"
)

var HTTPMiddlewareList []HttpMiddleware

func RegisterHTTPMiddleware(middleware HttpMiddleware) {
	HTTPMiddlewareList = append(HTTPMiddlewareList, middleware)
}

type HttpHandler func(writer http.ResponseWriter, request *http.Request)

type HttpMiddleware func(http.Handler) http.Handler

func Chain(middlewares ...HttpMiddleware) HttpMiddleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- { // reverse
			next = middlewares[i](next)
		}
		return next
	}
}

type BeforeHandler func(ctx context.Context, req interface{}) (context.Context, interface{}, error)
type AfterHandler func(response interface{}, err error) (interface{}, error)
