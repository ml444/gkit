// Package slog adapts a standard library slog.Logger to gkit's log.Logger.
//
// Usage:
//
//	import (
//	    "log/slog"
//
//	    "github.com/ml444/gkit/log"
//	    gkitslog "github.com/ml444/gkit/log/slog"
//	)
//
//	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
//	log.SetLogger(gkitslog.New(slog.New(h), gkitslog.WithSyncGkitLevel(true)))
//
// Note: Fatal logs call os.Exit(1) after writing at Error level, unlike gkit's DefaultLogger.
package slog

import (
	"fmt"
	"log/slog"
	"os"

	gkitlog "github.com/ml444/gkit/log"
	"github.com/ml444/gkit/log/internal/adapt"
)

type Option = adapt.Option

var WithSyncGkitLevel = adapt.WithSyncGkitLevel
var WithLoggerName = adapt.WithLoggerName

var exitFunc = os.Exit

type logger struct {
	adapt.Named
	l             *slog.Logger
	syncGkitLevel bool
}

func New(l *slog.Logger, opts ...Option) gkitlog.Logger {
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
		l.l.Error(fmt.Sprint(v...))
		exitFunc(1)
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
		l.l.Error(fmt.Sprintf(format, v...))
		exitFunc(1)
	}
}
