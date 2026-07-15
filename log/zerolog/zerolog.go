// Package zerolog adapts rs/zerolog to gkit's log.Logger.
//
// Usage:
//
//	import (
//	    "github.com/rs/zerolog"
//
//	    "github.com/ml444/gkit/log"
//	    gkitzerolog "github.com/ml444/gkit/log/zerolog"
//	)
//
//	zl := zerolog.New(os.Stdout).With().Timestamp().Logger()
//	log.SetLogger(gkitzerolog.New(zl))
//
// Note: Fatal logs terminate the process via zerolog's native Fatal behavior.
package zerolog

import (
	"fmt"

	gkitlog "github.com/ml444/gkit/log"
	"github.com/ml444/gkit/log/internal/adapt"
	"github.com/rs/zerolog"
)

type Option = adapt.Option

var WithSyncGkitLevel = adapt.WithSyncGkitLevel
var WithLoggerName = adapt.WithLoggerName

type logger struct {
	adapt.Named
	l             zerolog.Logger
	syncGkitLevel bool
}

func New(l zerolog.Logger, opts ...Option) gkitlog.Logger {
	o := adapt.Apply(opts...)
	return &logger{
		Named:         adapt.Named{Name: o.LoggerName},
		l:             l,
		syncGkitLevel: o.SyncGkitLevel,
	}
}

func (l *logger) Debug(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.DebugLevel) {
		l.l.Debug().Msg(fmt.Sprint(v...))
	}
}

func (l *logger) Info(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.InfoLevel) {
		l.l.Info().Msg(fmt.Sprint(v...))
	}
}

func (l *logger) Warn(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.WarnLevel) {
		l.l.Warn().Msg(fmt.Sprint(v...))
	}
}

func (l *logger) Error(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.ErrorLevel) {
		l.l.Error().Msg(fmt.Sprint(v...))
	}
}

func (l *logger) Fatal(v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.FatalLevel) {
		l.l.Fatal().Msg(fmt.Sprint(v...))
	}
}

func (l *logger) Printf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.InfoLevel) {
		l.l.Info().Msgf(format, v...)
	}
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.DebugLevel) {
		l.l.Debug().Msgf(format, v...)
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.InfoLevel) {
		l.l.Info().Msgf(format, v...)
	}
}

func (l *logger) Warnf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.WarnLevel) {
		l.l.Warn().Msgf(format, v...)
	}
}

func (l *logger) Errorf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.ErrorLevel) {
		l.l.Error().Msgf(format, v...)
	}
}

func (l *logger) Fatalf(format string, v ...interface{}) {
	if adapt.ShouldLog(l.syncGkitLevel, gkitlog.FatalLevel) {
		l.l.Fatal().Msgf(format, v...)
	}
}
