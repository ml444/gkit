package log

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Colors
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"
)

// NewGormLogger initialize logger
func NewGormLogger(cfg gormlogger.Config) gormlogger.Interface {
	var (
		infoStr      = "%s %s"
		warnStr      = "%s %s"
		errStr       = "%s %s"
		traceStr     = "[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s [%.3fms] [rows:%v] %s"
		traceErrStr  = "%s [%.3fms] [rows:%v] %s"
	)

	if cfg.Colorful {
		// infoStr = Green + "[INF] " + Reset + "%s %s"
		// warnStr = Magenta + "[WAR] " + Reset + "%s %s"
		// errStr = Red + "[ERR] " + Reset + "%s %s"
		traceStr = Yellow + " [%.3fms]" + BlueBold + " [rows:%v]" + Reset + " %s"
		traceWarnStr = Yellow + "%s" + Reset + RedBold + " [%.3fms]" + Yellow + " [rows:%v]" + Magenta + " %s" + Reset
		traceErrStr = MagentaBold + "%s" + Yellow + " [%.3fms]" + BlueBold + " [rows:%v]" + Reset + " %s"
	}

	return &dbLogger{
		Writer:       GetLogger(),
		Config:       cfg,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type dbLogger struct {
	// gormlogger.Writer
	Writer Logger
	gormlogger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

// LogMode log mode
func (l *dbLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *dbLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.Writer.Infof(l.infoStr, msg, fmt.Sprint(data...))
	}
}

// Warn print warn messages
func (l *dbLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.Writer.Warnf(l.warnStr, msg, fmt.Sprintln(data...))
	}
}

// Error print error messages
func (l *dbLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.Writer.Errorf(l.errStr, msg, fmt.Sprintln(data...))
	}
}

// Trace print sql message
func (l *dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Writer.Errorf(l.traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Writer.Errorf(l.traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Writer.Warnf(l.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Writer.Warnf(l.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormlogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Writer.Infof(l.traceStr, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Writer.Infof(l.traceStr, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// ParamsFilter Trace print sql message
func (l *dbLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
