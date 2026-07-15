package logrus

import (
	"io"
	"testing"

	gkitlog "github.com/ml444/gkit/log"
	"github.com/sirupsen/logrus"
)

type entryHook struct {
	entries []logrus.Entry
}

func (h *entryHook) Levels() []logrus.Level { return logrus.AllLevels }

func (h *entryHook) Fire(e *logrus.Entry) error {
	h.entries = append(h.entries, *e)
	return nil
}

func newTestLogger(t *testing.T) (*entryHook, gkitlog.Logger) {
	t.Helper()
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	l.SetOutput(io.Discard)
	h := &entryHook{}
	l.AddHook(h)
	return h, New(l)
}

func TestNewNil(t *testing.T) {
	if New(nil) == nil {
		t.Fatal("New(nil) should return nop logger, not nil")
	}
	if NewEntry(nil) == nil {
		t.Fatal("NewEntry(nil) should return nop logger, not nil")
	}
}

func TestInfoFormatting(t *testing.T) {
	h, lg := newTestLogger(t)
	lg.Info("a", "b")
	if len(h.entries) != 1 || h.entries[0].Message != "ab" {
		t.Fatalf("Info entries = %+v", h.entries)
	}
}

func TestInfofFormatting(t *testing.T) {
	h, lg := newTestLogger(t)
	lg.Infof("hi %s", "x")
	if len(h.entries) != 1 || h.entries[0].Message != "hi x" {
		t.Fatalf("Infof entries = %+v", h.entries)
	}
}

func TestLevelMapping(t *testing.T) {
	h, lg := newTestLogger(t)
	lg.Debug("d")
	lg.Warn("w")
	lg.Error("e")

	if len(h.entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(h.entries))
	}
	if h.entries[0].Level != logrus.DebugLevel || h.entries[1].Level != logrus.WarnLevel || h.entries[2].Level != logrus.ErrorLevel {
		t.Fatalf("unexpected levels: %+v", h.entries)
	}
}

func TestNewEntry(t *testing.T) {
	l := logrus.New()
	l.SetOutput(io.Discard)
	h := &entryHook{}
	l.AddHook(h)
	lg := NewEntry(l.WithField("k", "v"))
	lg.Info("ok")
	if len(h.entries) != 1 || h.entries[0].Message != "ok" {
		t.Fatalf("NewEntry logs = %+v", h.entries)
	}
}

func TestSyncGkitLevel(t *testing.T) {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)
	l.SetOutput(io.Discard)
	h := &entryHook{}
	l.AddHook(h)
	gkitlog.SetLogLevel(gkitlog.WarnLevel)
	lg := New(l, WithSyncGkitLevel(true))

	lg.Debug("hidden")
	lg.Info("hidden")
	lg.Warn("visible")

	if len(h.entries) != 1 || h.entries[0].Message != "visible" {
		t.Fatalf("sync filter entries = %+v", h.entries)
	}
}
