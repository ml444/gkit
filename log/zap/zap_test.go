package zap

import (
	"testing"

	gkitlog "github.com/ml444/gkit/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func newTestLogger(t *testing.T) (*observer.ObservedLogs, gkitlog.Logger) {
	t.Helper()
	core, logs := observer.New(zapcore.DebugLevel)
	return logs, New(zap.New(core))
}

func TestNewNil(t *testing.T) {
	if New(nil) == nil {
		t.Fatal("New(nil) should return nop logger, not nil")
	}
}

func TestInfoFormatting(t *testing.T) {
	logs, lg := newTestLogger(t)
	lg.Info("a", "b")
	if len(logs.All()) != 1 || logs.All()[0].Message != "ab" {
		t.Fatalf("Info logs = %+v, want message %q", logs.All(), "ab")
	}
}

func TestInfofFormatting(t *testing.T) {
	logs, lg := newTestLogger(t)
	lg.Infof("hi %s", "x")
	if len(logs.All()) != 1 || logs.All()[0].Message != "hi x" {
		t.Fatalf("Infof logs = %+v, want message %q", logs.All(), "hi x")
	}
}

func TestLevelMapping(t *testing.T) {
	logs, lg := newTestLogger(t)
	lg.Debug("d")
	lg.Warn("w")
	lg.Error("e")

	entries := logs.All()
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	if entries[0].Level != zapcore.DebugLevel || entries[1].Level != zapcore.WarnLevel || entries[2].Level != zapcore.ErrorLevel {
		t.Fatalf("unexpected levels: %+v", entries)
	}
}

func TestLoggerName(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	lg := New(zap.New(core), WithLoggerName("svc"))
	if got := lg.GetLoggerName(); got != "svc" {
		t.Fatalf("GetLoggerName() = %q, want svc", got)
	}
}

func TestSyncGkitLevel(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)
	gkitlog.SetLogLevel(gkitlog.WarnLevel)
	lg := New(zap.New(core), WithSyncGkitLevel(true))

	lg.Debug("hidden")
	lg.Info("hidden")
	lg.Warn("visible")

	if len(logs.All()) != 1 || logs.All()[0].Message != "visible" {
		t.Fatalf("sync filter logs = %+v", logs.All())
	}
}
