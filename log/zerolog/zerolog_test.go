package zerolog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	gkitlog "github.com/ml444/gkit/log"
	"github.com/rs/zerolog"
)

func newTestLogger(t *testing.T) (*bytes.Buffer, gkitlog.Logger) {
	t.Helper()
	var buf bytes.Buffer
	zl := zerolog.New(&buf)
	return &buf, New(zl)
}

func TestNewNilIsNop(t *testing.T) {
	lg := New(zerolog.Nop())
	lg.Info("discarded")
}

func TestInfoFormatting(t *testing.T) {
	buf, lg := newTestLogger(t)
	lg.Info("a", "b")

	var rec map[string]string
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec["message"] != "ab" {
		t.Fatalf("message = %q, want %q", rec["message"], "ab")
	}
}

func TestInfofFormatting(t *testing.T) {
	buf, lg := newTestLogger(t)
	lg.Infof("hi %s", "x")

	var rec map[string]string
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if rec["message"] != "hi x" {
		t.Fatalf("message = %q, want %q", rec["message"], "hi x")
	}
}

func TestLevelMapping(t *testing.T) {
	buf, lg := newTestLogger(t)
	lg.Debug("d")
	lg.Warn("w")
	lg.Error("e")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	for i, want := range []string{"debug", "warn", "error"} {
		var rec map[string]string
		if err := json.Unmarshal([]byte(lines[i]), &rec); err != nil {
			t.Fatalf("unmarshal line %d: %v", i, err)
		}
		if rec["level"] != want {
			t.Fatalf("line %d level = %q, want %q", i, rec["level"], want)
		}
	}
}

func TestSyncGkitLevel(t *testing.T) {
	var buf bytes.Buffer
	zl := zerolog.New(&buf)
	gkitlog.SetLogLevel(gkitlog.WarnLevel)
	lg := New(zl, WithSyncGkitLevel(true))

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
