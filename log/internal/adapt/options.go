package adapt

type Options struct {
	SyncGkitLevel bool
	LoggerName    string
}

type Option func(*Options)

func WithSyncGkitLevel(enable bool) Option {
	return func(o *Options) {
		o.SyncGkitLevel = enable
	}
}

func WithLoggerName(name string) Option {
	return func(o *Options) {
		o.LoggerName = name
	}
}

func Apply(opts ...Option) Options {
	var o Options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
