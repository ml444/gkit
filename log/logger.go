package log

import (
	"fmt"
	"os"
)

type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

func NewNopLogger() Logger {
	return nil
}

func NewStdLogger() Logger {
	return StdLogger{}
}

type StdLogger struct{}

func (_ StdLogger) Debug(values ...interface{}) { os.Stdout.Write([]byte(fmt.Sprint(values...))) }
func (_ StdLogger) Info(values ...interface{})  { os.Stdout.Write([]byte(fmt.Sprint(values...))) }
func (_ StdLogger) Warn(values ...interface{})  { os.Stdout.Write([]byte(fmt.Sprint(values...))) }
func (_ StdLogger) Error(values ...interface{}) { os.Stdout.Write([]byte(fmt.Sprint(values...))) }

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
