package coder

import (
	"sync"
	"testing"
)

// BenchmarkGetCoder compares the production lookup with the former RWMutex
// read path. Run with -run '^$' -bench BenchmarkGetCoder -benchmem -cpu=1,8
// to measure the built-in registry without registrations from unit tests.
func BenchmarkGetCoder(b *testing.B) {
	baseline := Snapshot()
	var mu sync.RWMutex
	lockedGet := func(name string) ICoder {
		mu.RLock()
		defer mu.RUnlock()
		if c, ok := baseline[name]; ok {
			return c
		}
		return baseline["json"]
	}
	for _, query := range []struct{ name, key string }{
		{"Hit", "json"}, {"Fallback", "benchmark-missing"},
	} {
		for _, method := range []struct {
			name string
			get  func(string) ICoder
		}{
			{"AtomicSnapshot", GetCoder}, {"RWMutexBaseline", lockedGet},
		} {
			b.Run(query.name+"/"+method.name, func(b *testing.B) {
				b.ReportAllocs()
				b.RunParallel(func(pb *testing.PB) {
					var result ICoder
					ran := false
					for pb.Next() {
						result = method.get(query.key)
						ran = true
					}
					if ran && result == nil {
						b.Error("missing codec")
					}
				})
			})
		}
	}
}
