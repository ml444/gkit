package log

import "sync/atomic"

var levelVal atomic.Int32

func SetLogLevel(lvl LogLevel) {
	levelVal.Store(int32(lvl))
}

// CurrentLevel returns the global log level set by SetLogLevel.
func CurrentLevel() LogLevel {
	return LogLevel(levelVal.Load())
}
