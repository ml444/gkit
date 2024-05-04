package middleware

import (
	"context"
	"net/http"
)

type (
	ServiceHandler func(ctx context.Context, req interface{}) (rsp interface{}, err error)
	BeforeHandler  func(ctx context.Context, req interface{}) (context.Context, interface{}, error)
	AfterHandler   func(rsp interface{}, err error) (interface{}, error)
)

type Middleware func(ServiceHandler) ServiceHandler

func Chain(middlewares ...Middleware) Middleware {
	return func(next ServiceHandler) ServiceHandler {
		for i := len(middlewares) - 1; i >= 0; i-- { // reverse
			next = middlewares[i](next)
		}
		return next
	}
}

/*
=============== HTTP =====================
*/

type HttpMiddleware func(http.Handler) http.Handler

func HTTPChain(middlewares ...HttpMiddleware) HttpMiddleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- { // reverse
			next = middlewares[i](next)
		}
		return next
	}
}
