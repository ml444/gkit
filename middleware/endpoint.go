package middleware

import "context"

// Endpoint is the fundamental building block of servers and clients.
// It represents a single RPC method.
type Endpoint func(ctx context.Context, request interface{}) (response interface{}, err error)

// Middleware is a chainable behavior modifier for endpoints.
type Middleware func(Endpoint) Endpoint

// LabeledMiddleware will get passed the endpoint name when passed to
// WrapAllLabeledExcept, this can be used to write a generic metrics
// middleware which can send the endpoint name to the metrics collector.
type LabeledMiddleware func(string, Endpoint) Endpoint

func Chain(m ...Middleware) Middleware {
	return func(next Endpoint) Endpoint {
		for i := len(m) - 1; i >= 0; i-- { // reverse
			next = m[i](next)
		}
		return next
	}
}
