package middleware

import "context"

var (
	defaultOptions = &Options{
		HandlerFunc: nil,
	}
)

type HandlerFunc func(p any) (err error)

type HandlerFuncContext func(ctx context.Context, p any) (err error)

func WrapHandlerFuncContext(f HandlerFunc) HandlerFuncContext {
	return func(ctx context.Context, p any) (err error) {
		return f(p)
	}
}

type Options struct {
	HandlerFunc HandlerFuncContext
}

func EvaluateOptions(opts []Option) *Options {
	optCopy := &Options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*Options)
