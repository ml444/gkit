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

type Logger interface {
	GetLoggerName() string
	SetLoggerName(string)

	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Fatal(...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
}

func NewNopLogger() Logger {
	return nil
}

func SetLogger(l Logger) {
	logger = l
}

func NewDefaultLogger(output io.Writer) Logger {
	return &DefaultLogger{writer: output}
}

type DefaultLogger struct {
	name   string
	writer io.Writer
}

func (l *DefaultLogger) Log(level string, msg string) {
	l.writer.Write([]byte("["))
	l.writer.Write([]byte(level))
	l.writer.Write([]byte("] "))
	l.writer.Write([]byte(msg))
}
func (l DefaultLogger) GetLoggerName() string {
	return l.name
}
func (l DefaultLogger) SetLoggerName(name string) {
	l.name = name
}
func (l *DefaultLogger) Debug(values ...interface{}) { l.Log("DEG", fmt.Sprint(values...)) }
func (l *DefaultLogger) Info(values ...interface{})  { l.Log("INF", fmt.Sprint(values...)) }
func (l *DefaultLogger) Warn(values ...interface{})  { l.Log("WAN", fmt.Sprint(values...)) }
func (l *DefaultLogger) Error(values ...interface{}) { l.Log("ERR", fmt.Sprint(values...)) }
func (l *DefaultLogger) Fatal(values ...interface{}) { l.Log("FAT", fmt.Sprint(values...)) }

func (l *DefaultLogger) Debugf(template string, values ...interface{}) {
	l.Log("DEG", fmt.Sprintf(template, values...))
}
func (l *DefaultLogger) Infof(template string, values ...interface{}) {
	l.Log("INF", fmt.Sprintf(template, values...))
}
func (l *DefaultLogger) Warnf(template string, values ...interface{}) {
	l.Log("WAN", fmt.Sprintf(template, values...))
}
func (l *DefaultLogger) Errorf(template string, values ...interface{}) {
	l.Log("ERR", fmt.Sprintf(template, values...))
}
func (l *DefaultLogger) Fatalf(template string, values ...interface{}) {
	l.Log("FAT", fmt.Sprintf(template, values...))
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
