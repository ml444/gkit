package log

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
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
		infoStr      = "[INF] %s %s"
		warnStr      = "[WAR] %s %s"
		errStr       = "[ERR] %s %s"
		traceStr     = "[INF] [%.3fms] [rows:%v] %s\n"
		traceWarnStr = "[WAR] %s [%.3fms] [rows:%v] %s\n"
		traceErrStr  = "[ERR] %s [%.3fms] [rows:%v] %s\n"
	)

	if cfg.Colorful {
		infoStr = Green + "[INF] " + Reset + "%s %s"
		warnStr = Magenta + "[WAR] " + Reset + "%s %s"
		errStr = Red + "[ERR] " + Reset + "%s %s"
		traceStr = Green + "[INF] " + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s\n"
		traceWarnStr = Magenta + "[WAR] " + Yellow + "%s" + Reset + RedBold + "[%.3fms] " + Yellow + "[rows:%v]" + Magenta + " %s" + Reset + "\n"
		traceErrStr = Red + "[ERR] " + MagentaBold + "%s" + Yellow + "[%.3fms] " + BlueBold + "[rows:%v]" + Reset + " %s\n"
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
	gormlogger.Writer
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
func (l dbLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.Printf(l.infoStr, msg, fmt.Sprintln(data...))
	}
}

// Warn print warn messages
func (l dbLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.Printf(l.warnStr, msg, fmt.Sprintln(data...))
	}
}

// Error print error messages
func (l dbLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.Printf(l.errStr, msg, fmt.Sprintln(data...))
	}
}

// Trace print sql message
func (l dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceErrStr, err, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Printf(l.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceWarnStr, slowLog, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	case l.LogLevel == gormlogger.Info:
		sql, rows := fc()
		if rows == -1 {
			l.Printf(l.traceStr, float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			l.Printf(l.traceStr, float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}

// ParamsFilter Trace print sql message
func (l dbLogger) ParamsFilter(ctx context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
