package glog

import (
	"fmt"
	"testing"

	gkitlog "github.com/ml444/gkit/log"
)

type stubLogger struct {
	name string
	logs []string
}

func (s *stubLogger) GetLoggerName() string              { return s.name }
func (s *stubLogger) SetLoggerName(name string)          { s.name = name }
func (s *stubLogger) Debug(v ...interface{})             { s.logs = append(s.logs, "debug:"+fmt.Sprint(v...)) }
func (s *stubLogger) Info(v ...interface{})              { s.logs = append(s.logs, "info:"+fmt.Sprint(v...)) }
func (s *stubLogger) Warn(v ...interface{})              { s.logs = append(s.logs, "warn:"+fmt.Sprint(v...)) }
func (s *stubLogger) Error(v ...interface{})             { s.logs = append(s.logs, "error:"+fmt.Sprint(v...)) }
func (s *stubLogger) Fatal(v ...interface{})             { s.logs = append(s.logs, "fatal:"+fmt.Sprint(v...)) }
func (s *stubLogger) Debugf(format string, v ...interface{}) {
	s.logs = append(s.logs, "debugf:"+fmt.Sprintf(format, v...))
}
func (s *stubLogger) Infof(format string, v ...interface{}) {
	s.logs = append(s.logs, "infof:"+fmt.Sprintf(format, v...))
}
func (s *stubLogger) Warnf(format string, v ...interface{}) {
	s.logs = append(s.logs, "warnf:"+fmt.Sprintf(format, v...))
}
func (s *stubLogger) Errorf(format string, v ...interface{}) {
	s.logs = append(s.logs, "errorf:"+fmt.Sprintf(format, v...))
}
func (s *stubLogger) Fatalf(format string, v ...interface{}) {
	s.logs = append(s.logs, "fatalf:"+fmt.Sprintf(format, v...))
}
func (s *stubLogger) Printf(format string, v ...interface{}) {
	s.logs = append(s.logs, "printf:"+fmt.Sprintf(format, v...))
}

func TestNewNil(t *testing.T) {
	if New(nil) == nil {
		t.Fatal("New(nil) should return nop logger, not nil")
	}
}

func TestInfoFormatting(t *testing.T) {
	stub := &stubLogger{}
	lg := newLogger(stub)
	lg.Info("a", "b")
	if len(stub.logs) != 1 || stub.logs[0] != "info:ab" {
		t.Fatalf("logs = %+v", stub.logs)
	}
}

func TestInfofFormatting(t *testing.T) {
	stub := &stubLogger{}
	lg := newLogger(stub)
	lg.Infof("hi %s", "x")
	if len(stub.logs) != 1 || stub.logs[0] != "infof:hi x" {
		t.Fatalf("logs = %+v", stub.logs)
	}
}

func TestPrintfUsesInfof(t *testing.T) {
	stub := &stubLogger{}
	lg := newLogger(stub)
	lg.Printf("hello")
	if len(stub.logs) != 1 || stub.logs[0] != "infof:hello" {
		t.Fatalf("logs = %+v", stub.logs)
	}
}

func TestSyncGkitLevel(t *testing.T) {
	stub := &stubLogger{}
	gkitlog.SetLogLevel(gkitlog.WarnLevel)
	lg := newLogger(stub, WithSyncGkitLevel(true))

	lg.Debug("hidden")
	lg.Info("hidden")
	lg.Warn("visible")

	if len(stub.logs) != 1 || stub.logs[0] != "warn:visible" {
		t.Fatalf("sync filter logs = %+v", stub.logs)
	}
}

func TestLoggerName(t *testing.T) {
	stub := &stubLogger{}
	lg := newLogger(stub, WithLoggerName("svc"))
	if got := lg.GetLoggerName(); got != "svc" {
		t.Fatalf("GetLoggerName() = %q, want svc", got)
	}
}
