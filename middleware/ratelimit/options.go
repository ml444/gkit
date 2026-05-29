package ratelimit

// Options configures rate limit middleware behavior.
type Options struct {
	// FailOpen when true allows requests if transport context is missing (default).
	FailOpen bool
	// Store overrides the default in-process limiter (e.g. Redis).
	Store Store
	// ServiceName is used in distributed store keys: gkit:rl:{service}:{path}:{window}.
	ServiceName string
}

// Option configures Options.
type Option func(*Options)

// WithFailClosed rejects requests when transport context is missing.
func WithFailClosed() Option {
	return func(o *Options) { o.FailOpen = false }
}

// WithStore sets a custom rate limit store.
func WithStore(store Store) Option {
	return func(o *Options) { o.Store = store }
}

// WithServiceName sets the service segment in Redis keys (gkit:rl:{service}:...).
func WithServiceName(name string) Option {
	return func(o *Options) { o.ServiceName = name }
}

func applyOptions(opts []Option) Options {
	o := Options{FailOpen: true}
	for _, fn := range opts {
		fn(&o)
	}
	return o
}
