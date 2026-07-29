package log

import "sync/atomic"

type LogLevel int

const (
	DebugLevel LogLevel = iota + 1
	PrintLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

var levelVal atomic.Int32

func SetLogLevel(lvl LogLevel) {
	levelVal.Store(int32(lvl))
}

// CurrentLevel returns the global log level set by SetLogLevel.
func CurrentLevel() LogLevel {
	return LogLevel(levelVal.Load())
}
