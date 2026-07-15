package slog

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	gkitlog "github.com/ml444/gkit/log"
)

func newTestLogger(t *testing.T) (*bytes.Buffer, gkitlog.Logger) {
	t.Helper()
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	return &buf, New(slog.New(h))
}

func TestNewNil(t *testing.T) {
	if New(nil) == nil {
		t.Fatal("New(nil) should return nop logger, not nil")
	}
}

func TestInfoFormatting(t *testing.T) {
	buf, lg := newTestLogger(t)
	lg.Info("a", "b")
	if !strings.Contains(buf.String(), "ab") {
		t.Fatalf("Info output = %q, want message containing %q", buf.String(), "ab")
	}
}

func TestInfofFormatting(t *testing.T) {
	buf, lg := newTestLogger(t)
	lg.Infof("hi %s", "x")
	if !strings.Contains(buf.String(), "hi x") {
		t.Fatalf("Infof output = %q, want %q", buf.String(), "hi x")
	}
}

func TestLoggerName(t *testing.T) {
	_, lg := newTestLogger(t)
	lg = New(slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil)), WithLoggerName("svc"))
	if got := lg.GetLoggerName(); got != "svc" {
		t.Fatalf("GetLoggerName() = %q, want svc", got)
	}
	lg.SetLoggerName("other")
	if got := lg.GetLoggerName(); got != "other" {
		t.Fatalf("GetLoggerName() after Set = %q, want other", got)
	}
}

func TestSyncGkitLevel(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	gkitlog.SetLogLevel(gkitlog.WarnLevel)
	lg := New(slog.New(h), WithSyncGkitLevel(true))

	lg.Debug("hidden")
	lg.Info("hidden")
	lg.Warn("visible")

	out := buf.String()
	if strings.Contains(out, "hidden") {
		t.Fatalf("sync filter should drop debug/info, got %q", out)
	}
	if !strings.Contains(out, "visible") {
		t.Fatalf("sync filter should allow warn, got %q", out)
	}
}

func TestFatalDoesNotExitWhenFiltered(t *testing.T) {
	exitCalled := false
	oldExit := exitFunc
	exitFunc = func(code int) { exitCalled = true }
	t.Cleanup(func() { exitFunc = oldExit })

	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	lg := New(slog.New(h), WithSyncGkitLevel(true))
	gkitlog.SetLogLevel(gkitlog.PanicLevel)

	lg.Fatal("should not exit")
	if exitCalled {
		t.Fatal("Fatal should not exit when filtered by gkit level")
	}
}
