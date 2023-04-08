package log

import (
	"fmt"
	"os"
)

func init() {
	if logger == nil {
		logger = NewStdLogger()
	}
}

var logger Logger

type Logger interface {
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

func NewStdLogger() Logger {
	return StdLogger{}
}

type StdLogger struct{}

func (_ StdLogger) Debug(values ...interface{}) { os.Stdout.Write([]byte(fmt.Sprint(values...))) }
func (_ StdLogger) Info(values ...interface{})  { os.Stdout.Write([]byte(fmt.Sprint(values...))) }
func (_ StdLogger) Warn(values ...interface{})  { os.Stdout.Write([]byte(fmt.Sprint(values...))) }
func (_ StdLogger) Error(values ...interface{}) { os.Stdout.Write([]byte(fmt.Sprint(values...))) }
func (_ StdLogger) Fatal(values ...interface{}) { os.Stdout.Write([]byte(fmt.Sprint(values...))) }

func (_ StdLogger) Debugf(template string, values ...interface{}) {
	os.Stdout.Write([]byte(fmt.Sprintf(template, values...)))
}
func (_ StdLogger) Infof(template string, values ...interface{}) {
	os.Stdout.Write([]byte(fmt.Sprintf(template, values...)))
}
func (_ StdLogger) Warnf(template string, values ...interface{}) {
	os.Stdout.Write([]byte(fmt.Sprintf(template, values...)))
}
func (_ StdLogger) Errorf(template string, values ...interface{}) {
	os.Stdout.Write([]byte(fmt.Sprintf(template, values...)))
}
func (_ StdLogger) Fatalf(template string, values ...interface{}) {
	os.Stdout.Write([]byte(fmt.Sprintf(template, values...)))
}

func Debug(values ...interface{}) { logger.Debug(values...) }
func Info(values ...interface{})  { logger.Info(values...) }
func Warn(values ...interface{})  { logger.Warn(values...) }
func Error(values ...interface{}) { logger.Error(values...) }
func Fatal(values ...interface{}) { logger.Error(values...) } // In order to print the stack log

func Debugf(template string, values ...interface{}) { logger.Debugf(template, values...) }
func Infof(template string, values ...interface{})  { logger.Infof(template, values...) }
func Warnf(template string, values ...interface{})  { logger.Warnf(template, values...) }
func Errorf(template string, values ...interface{}) { logger.Errorf(template, values...) }
func Fatalf(template string, values ...interface{}) { logger.Errorf(template, values...) }
