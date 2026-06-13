package log

import (
	"io"
	"sync"
	"testing"
)

// discardWriter is a no-op Writer used so the race test does not spam stdout.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error)       { return io.Discard.Write(p) }
func (discardWriter) WriteString(s string) (int, error) { return len(s), nil }

// TestGlobalLoggerNoRace exercises the lock-free global logger/level state under
// heavy concurrent reads (logging, GetLogger, getLevel) and writes
// (SetLogger, SetLogLevel). Run with `go test -race ./log/` to detect any data
// race on the global state.
func TestGlobalLoggerNoRace(t *testing.T) {
	// Restore the default logger/level after the test mutates global state.
	orig := GetLogger()
	origLevel := getLevel()
	t.Cleanup(func() {
		SetLogger(orig)
		SetLogLevel(origLevel)
	})

	SetLogger(NewDefaultLogger(discardWriter{}))

	const (
		readers        = 50
		writers        = 10
		iterPerReader  = 1000
		iterPerWriter  = 1000
		levelVariation = 7
	)

	var wg sync.WaitGroup

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterPerReader; j++ {
				Infof("hello %d", j)
				Debug("x")
				_ = GetLogger()
				_ = getLevel()
			}
		}()
	}

	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterPerWriter; j++ {
				SetLogLevel(LogLevel((j % levelVariation) + 1))
				if j%100 == 0 {
					SetLogger(NewDefaultLogger(discardWriter{}))
				}
			}
		}()
	}

	wg.Wait()
}
