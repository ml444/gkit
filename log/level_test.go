package log

import "testing"

func TestCurrentLevel(t *testing.T) {
	SetLogLevel(InfoLevel)
	if got := CurrentLevel(); got != InfoLevel {
		t.Fatalf("CurrentLevel() = %v, want InfoLevel", got)
	}

	SetLogLevel(DebugLevel)
	if got := CurrentLevel(); got != DebugLevel {
		t.Fatalf("CurrentLevel() = %v, want DebugLevel", got)
	}
}
