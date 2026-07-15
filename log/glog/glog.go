// Package glog adapts github.com/ml444/glog to gkit's log.Logger.
//
// Usage:
//
//	import (
//	    "github.com/ml444/gkit/log"
//	    gkitglog "github.com/ml444/gkit/log/glog"
//	    glog "github.com/ml444/glog"
//	)
//
//	lg, err := gkitglog.NewFromInit(
//	    glog.SetLoggerName("service"),
//	    glog.SetWorkerConfigs(glog.NewDefaultStdoutWorkerConfig()),
//	)
//	if err != nil {
//	    panic(err)
//	}
//	log.SetLogger(lg)
//
// Or wrap an existing glog logger:
//
//	log.SetLogger(gkitglog.New(glog.GetLogger()))
//
// Note: Fatal logs follow glog's native behavior and may terminate the process.
package glog

import (
	"fmt"

	gkitlog "github.com/ml444/gkit/log"
	"github.com/ml444/gkit/log/internal/adapt"
	glog "github.com/ml444/glog"
)

type Option = adapt.Option

var WithSyncGkitLevel = adapt.WithSyncGkitLevel
var WithLoggerName = adapt.WithLoggerName

type underlying interface {
	GetLoggerName() string
	SetLoggerName(string)

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
	Printf(string, ...interface{})
}

type logger struct {
	adapt.Named
	l             underlying
	syncGkitLevel bool
}

func New(l glog.ILogger, opts ...Option) gkitlog.Logger {
	if l == nil {
		return gkitlog.NewNopLogger()
	}
	return newLogger(l, opts...)
}

func NewFromInit(opts ...glog.OptionFunc) (gkitlog.Logger, error) {
	if err := glog.InitLog(opts...); err != nil {
		return nil, err
	}
	return New(glog.GetLogger()), nil
}

func newLogger(l underlying, opts ...Option) gkitlog.Logger {
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
