package log

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"
)

func init() {
	loggerVal.Store(loggerHolder{NewDefaultLogger(os.Stdout)})
}

// loggerHolder wraps the Logger interface in a single concrete type so it can
// be stored in an atomic.Value (which requires a consistent concrete type).
type loggerHolder struct {
	Logger
}

// Lock-free global state: reads (the hot path) never take a lock.
//   - loggerVal holds the active Logger (read-mostly, swapped via SetLogger).
//   - levelVal holds the current LogLevel as an int32.
var (
	loggerVal atomic.Value // stores loggerHolder
	levelVal  atomic.Int32
)

type Logger interface {
	GetLoggerName() string
	SetLoggerName(string)

	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Fatal(...interface{})

	Printf(template string, args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

type Writer interface {
	io.Writer
	io.StringWriter
}

// NewNopLogger returns a Logger that discards everything.
// It is a real no-op implementation (not nil) so callers never panic.
func NewNopLogger() Logger {
	return nopLoggerInstance
}

func SetLogger(lg Logger) {
	loggerVal.Store(loggerHolder{lg})
}
func GetLogger() Logger {
	v := loggerVal.Load()
	if v == nil {
		return nil
	}
	return v.(loggerHolder).Logger
}

// currentLogger returns the active logger, falling back to a no-op logger
// if it is unset or was explicitly set to nil, so package-level helpers never panic.
func currentLogger() Logger {
	if lg := GetLogger(); lg != nil {
		return lg
	}
	return nopLoggerInstance
}

func NewDefaultLogger(output Writer) *DefaultLogger {
	return &DefaultLogger{writer: output}
}

type DefaultLogger struct {
	name   string
	writer Writer
}

func (l *DefaultLogger) Log(lvl LogLevel, value string) {
	if lvl < getLevel() {
		return
	}
	_, _ = l.writer.WriteString(ColorLevel(lvl))
	_, _ = l.writer.WriteString(value + "\n")
}
func (l *DefaultLogger) GetLoggerName() string {
	return l.name
}
func (l *DefaultLogger) SetLoggerName(name string) {
	l.name = name
}
func (l *DefaultLogger) Debug(values ...interface{}) {
	l.Log(DebugLevel, fmt.Sprintln(values...))
}
func (l *DefaultLogger) Info(values ...interface{}) {
	l.Log(InfoLevel, fmt.Sprintln(values...))
}
func (l *DefaultLogger) Warn(values ...interface{}) {
	l.Log(WarnLevel, fmt.Sprintln(values...))
}
func (l *DefaultLogger) Error(values ...interface{}) {
	l.Log(ErrorLevel, fmt.Sprintln(values...))
}
func (l *DefaultLogger) Fatal(values ...interface{}) {
	l.Log(FatalLevel, fmt.Sprintln(values...))
}

func (l *DefaultLogger) Printf(template string, values ...interface{}) {
	_, err := l.writer.WriteString(fmt.Sprintf(template, values...))
	if err != nil {
		println(err.Error())
	}
}
func (l *DefaultLogger) Debugf(template string, values ...interface{}) {
	l.Log(DebugLevel, fmt.Sprintf(template, values...))
}
func (l *DefaultLogger) Infof(template string, values ...interface{}) {
	l.Log(InfoLevel, fmt.Sprintf(template, values...))
}
func (l *DefaultLogger) Warnf(template string, values ...interface{}) {
	l.Log(WarnLevel, fmt.Sprintf(template, values...))
}
func (l *DefaultLogger) Errorf(template string, values ...interface{}) {
	l.Log(ErrorLevel, fmt.Sprintf(template, values...))
}
func (l *DefaultLogger) Fatalf(template string, values ...interface{}) {
	l.Log(FatalLevel, fmt.Sprintf(template, values...))
}

var nopLoggerInstance Logger = (*nopLogger)(nil)

// nopLogger discards all log output.
type nopLogger struct{ name string }

func (l *nopLogger) GetLoggerName() string         { return "" }
func (l *nopLogger) SetLoggerName(string)          {}
func (l *nopLogger) Debug(...interface{})          {}
func (l *nopLogger) Info(...interface{})           {}
func (l *nopLogger) Warn(...interface{})           {}
func (l *nopLogger) Error(...interface{})          {}
func (l *nopLogger) Fatal(...interface{})          {}
func (l *nopLogger) Printf(string, ...interface{}) {}
func (l *nopLogger) Debugf(string, ...interface{}) {}
func (l *nopLogger) Infof(string, ...interface{})  {}
func (l *nopLogger) Warnf(string, ...interface{})  {}
func (l *nopLogger) Errorf(string, ...interface{}) {}
func (l *nopLogger) Fatalf(string, ...interface{}) {}

func SetLogLevel(lvl LogLevel) {
	levelVal.Store(int32(lvl))
}

func getLevel() LogLevel {
	return LogLevel(levelVal.Load())
}

func GetLoggerName() string {
	return currentLogger().GetLoggerName()
}
func SetLoggerName(name string) {
	currentLogger().SetLoggerName(name)
}
func Debug(values ...interface{}) { currentLogger().Debug(values...) }
func Info(values ...interface{})  { currentLogger().Info(values...) }
func Warn(values ...interface{})  { currentLogger().Warn(values...) }
func Error(values ...interface{}) { currentLogger().Error(values...) }
func Fatal(values ...interface{}) { currentLogger().Fatal(values...) } // In order to print the stack log

func Debugf(template string, values ...interface{}) { currentLogger().Debugf(template, values...) }
func Infof(template string, values ...interface{})  { currentLogger().Infof(template, values...) }
func Warnf(template string, values ...interface{})  { currentLogger().Warnf(template, values...) }
func Errorf(template string, values ...interface{}) { currentLogger().Errorf(template, values...) }
func Fatalf(template string, values ...interface{}) { currentLogger().Fatalf(template, values...) } // In order to print the stack log
