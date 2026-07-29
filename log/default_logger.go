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
	loggerVal  atomic.Value // stores loggerHolder
	enterBytes = []byte("\n")
)

type Logger interface {
	GetLoggerName() string
	SetLoggerName(string)

	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Fatal(...any)

	Printf(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
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
	return &DefaultLogger{writer: output, AutoEnter: true}
}

type DefaultLogger struct {
	name      string
	writer    Writer
	AutoEnter bool
}

func (l *DefaultLogger) Log(lvl LogLevel, values ...any) {
	if lvl < CurrentLevel() {
		return
	}
	_, _ = l.writer.WriteString(ColorLevel(lvl))
	_, _ = fmt.Fprintln(l.writer, values...)
}

func (l *DefaultLogger) Logf(lvl LogLevel, format string, values ...any) {
	if lvl < CurrentLevel() {
		return
	}
	_, _ = l.writer.WriteString(ColorLevel(lvl))
	_, _ = fmt.Fprintf(l.writer, format, values...)
	if l.AutoEnter {
		_, _ = l.writer.Write(enterBytes)
	}
}

func (l *DefaultLogger) GetLoggerName() string {
	return l.name
}
func (l *DefaultLogger) SetLoggerName(name string) {
	l.name = name
}
func (l *DefaultLogger) Debug(values ...any) {
	l.Log(DebugLevel, values...)
}
func (l *DefaultLogger) Info(values ...any) {
	l.Log(InfoLevel, values...)
}
func (l *DefaultLogger) Warn(values ...any) {
	l.Log(WarnLevel, values...)
}
func (l *DefaultLogger) Error(values ...any) {
	l.Log(ErrorLevel, values...)
}
func (l *DefaultLogger) Fatal(values ...any) {
	l.Log(FatalLevel, values...)
}

func (l *DefaultLogger) Debugf(format string, values ...any) {
	l.Logf(DebugLevel, format, values...)
}
func (l *DefaultLogger) Printf(format string, values ...any) {
	l.Logf(PrintLevel, format, values...)
}
func (l *DefaultLogger) Infof(format string, values ...any) {
	l.Logf(InfoLevel, format, values...)
}
func (l *DefaultLogger) Warnf(format string, values ...any) {
	l.Logf(WarnLevel, format, values...)
}
func (l *DefaultLogger) Errorf(format string, values ...any) {
	l.Logf(ErrorLevel, format, values...)
}
func (l *DefaultLogger) Fatalf(format string, values ...any) {
	l.Logf(FatalLevel, format, values...)
}

var nopLoggerInstance Logger = (*nopLogger)(nil)

// nopLogger discards all log output.
type nopLogger struct{ name string }

func (l *nopLogger) GetLoggerName() string { return "" }
func (l *nopLogger) SetLoggerName(string)  {}
func (l *nopLogger) Debug(...any)          {}
func (l *nopLogger) Info(...any)           {}
func (l *nopLogger) Warn(...any)           {}
func (l *nopLogger) Error(...any)          {}
func (l *nopLogger) Fatal(...any)          {}
func (l *nopLogger) Printf(string, ...any) {}
func (l *nopLogger) Debugf(string, ...any) {}
func (l *nopLogger) Infof(string, ...any)  {}
func (l *nopLogger) Warnf(string, ...any)  {}
func (l *nopLogger) Errorf(string, ...any) {}
func (l *nopLogger) Fatalf(string, ...any) {}

func GetLoggerName() string {
	return currentLogger().GetLoggerName()
}
func SetLoggerName(name string) {
	currentLogger().SetLoggerName(name)
}
func Debug(values ...any) { currentLogger().Debug(values...) }
func Info(values ...any)  { currentLogger().Info(values...) }
func Warn(values ...any)  { currentLogger().Warn(values...) }
func Error(values ...any) { currentLogger().Error(values...) }
func Fatal(values ...any) { currentLogger().Fatal(values...) } // In order to print the stack log

func Debugf(format string, values ...any) { currentLogger().Debugf(format, values...) }
func Infof(format string, values ...any)  { currentLogger().Infof(format, values...) }
func Warnf(format string, values ...any)  { currentLogger().Warnf(format, values...) }
func Errorf(format string, values ...any) { currentLogger().Errorf(format, values...) }
func Fatalf(format string, values ...any) { currentLogger().Fatalf(format, values...) } // In order to print the stack log
