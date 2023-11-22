package log

import (
	"fmt"
	"io"
	"os"
)

func init() {
	if logger == nil {
		logger = NewDefaultLogger(os.Stdout)
	}
}

var logger Logger
var level LogLevel

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

func NewNopLogger() Logger {
	return nil
}

func SetLogger(lg Logger) {
	logger = lg
}
func GetLogger() Logger {
	return logger
}

func NewDefaultLogger(output Writer) *DefaultLogger {
	return &DefaultLogger{writer: output}
}

type DefaultLogger struct {
	name   string
	writer Writer
}

func (l *DefaultLogger) Log(lvl LogLevel, value string) {
	if lvl < level {
		return
	}
	_, _ = l.writer.WriteString(ColorLevel(lvl))
	_, _ = l.writer.WriteString(value)
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
	println(err)
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

func SetLogLevel(lvl LogLevel) {
	level = lvl
}

func GetLoggerName() string {
	return logger.GetLoggerName()
}
func SetLoggerName(name string) {
	logger.SetLoggerName(name)
}
func Debug(values ...interface{}) { logger.Debug(values...) }
func Info(values ...interface{})  { logger.Info(values...) }
func Warn(values ...interface{})  { logger.Warn(values...) }
func Error(values ...interface{}) { logger.Error(values...) }
func Fatal(values ...interface{}) { logger.Fatal(values...) } // In order to print the stack log

func Debugf(template string, values ...interface{}) { logger.Debugf(template, values...) }
func Infof(template string, values ...interface{})  { logger.Infof(template, values...) }
func Warnf(template string, values ...interface{})  { logger.Warnf(template, values...) }
func Errorf(template string, values ...interface{}) { logger.Errorf(template, values...) }
func Fatalf(template string, values ...interface{}) { logger.Fatalf(template, values...) } // In order to print the stack log
