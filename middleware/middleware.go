package middleware

import (
	"context"
)

type ServiceHandler func(ctx context.Context, req interface{}) (interface{}, error)

type Middleware func(ServiceHandler) ServiceHandler

func Chain(middlewares ...Middleware) Middleware {
	return func(next ServiceHandler) ServiceHandler {
		for i := len(middlewares) - 1; i >= 0; i-- { // reverse
			next = middlewares[i](next)
		}
		return next
	}
}
