// Package logrus adapts sirupsen/logrus to gkit's log.Logger.
//
// Usage:
//
//	import (
//	    "github.com/sirupsen/logrus"
//
//	    "github.com/ml444/gkit/log"
//	    gkitlogrus "github.com/ml444/gkit/log/logrus"
//	)
//
//	l := logrus.New()
//	log.SetLogger(gkitlogrus.New(l))
//
// Note: Fatal logs terminate the process via logrus's native Fatal behavior.
package logrus

import (
	"fmt"

	gkitlog "github.com/ml444/gkit/log"
	"github.com/ml444/gkit/log/internal/adapt"
	"github.com/sirupsen/logrus"
)

type Option = adapt.Option

var WithSyncGkitLevel = adapt.WithSyncGkitLevel
var WithLoggerName = adapt.WithLoggerName

type leveledLogger interface {
	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{})
	Fatal(...interface{})

	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
}

type logger struct {
	adapt.Named
	l             leveledLogger
	syncGkitLevel bool
}

func New(l *logrus.Logger, opts ...Option) gkitlog.Logger {
	if l == nil {
		return gkitlog.NewNopLogger()
	}
	return newLogger(l, opts...)
}

func NewEntry(e *logrus.Entry, opts ...Option) gkitlog.Logger {
	if e == nil {
		return gkitlog.NewNopLogger()
	}
	return newLogger(e, opts...)
}

func newLogger(l leveledLogger, opts ...Option) gkitlog.Logger {
	o := adapt.Apply(opts...)
	return &logger{
		Named:         adapt.Named{Name: o.LoggerName},
		l:             l,
		syncGkitLevel: o.SyncGkitLevel,
	}
}

func (l *logger) Debug(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.DebugLevel) {
		l.l.Debug(fmt.Sprint(v...))
	}
}

func (l *logger) Info(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.InfoLevel) {
		l.l.Info(fmt.Sprint(v...))
	}
}

func (l *logger) Warn(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.WarnLevel) {
		l.l.Warn(fmt.Sprint(v...))
	}
}

func (l *logger) Error(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.ErrorLevel) {
		l.l.Error(fmt.Sprint(v...))
	}
}

func (l *logger) Fatal(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.FatalLevel) {
		l.l.Fatal(fmt.Sprint(v...))
	}
}

func (l *logger) Printf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.InfoLevel) {
		l.l.Infof(format, v...)
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.DebugLevel) {
		l.l.Debugf(format, v...)
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.InfoLevel) {
		l.l.Infof(format, v...)
	}
}

func (l *logger) Warnf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.WarnLevel) {
		l.l.Warnf(format, v...)
	}
}

func (l *logger) Errorf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.ErrorLevel) {
		l.l.Errorf(format, v...)
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.FatalLevel) {
		l.l.Fatalf(format, v...)
	}
}
