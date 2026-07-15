// Package zap adapts go.uber.org/zap.Logger to gkit's log.Logger.
//
// Usage:
//
//	import (
//	    "go.uber.org/zap"
//
//	    "github.com/ml444/gkit/log"
//	    gkitzap "github.com/ml444/gkit/log/zap"
//	)
//
//	z, _ := zap.NewProduction()
//	log.SetLogger(gkitzap.New(z))
//
// Note: Fatal logs terminate the process via zap's native Fatal behavior.
package zap

import (
	"fmt"

	gkitlog "github.com/ml444/gkit/log"
	"github.com/ml444/gkit/log/internal/adapt"
	"go.uber.org/zap"
)

type Option = adapt.Option

var WithSyncGkitLevel = adapt.WithSyncGkitLevel
var WithLoggerName = adapt.WithLoggerName

type logger struct {
	adapt.Named
	l             *zap.Logger
	syncGkitLevel bool
}

func New(l *zap.Logger, opts ...Option) gkitlog.Logger {
	if l == nil {
		return gkitlog.NewNopLogger()
	}
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
		l.l.Info(fmt.Sprintf(format, v...))
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.DebugLevel) {
		l.l.Debug(fmt.Sprintf(format, v...))
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.InfoLevel) {
		l.l.Info(fmt.Sprintf(format, v...))
	}
}

func (l *logger) Warnf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.WarnLevel) {
		l.l.Warn(fmt.Sprintf(format, v...))
	}
}

func (l *logger) Errorf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.ErrorLevel) {
		l.l.Error(fmt.Sprintf(format, v...))
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.FatalLevel) {
		l.l.Fatal(fmt.Sprintf(format, v...))
	}
}
