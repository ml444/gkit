package transport

import (
	"fmt"
	"strings"
	"testing"
)

var benchmarkMetadata MD

func BenchmarkMetadataNew(b *testing.B) {
	for _, count := range []int{0, 8, 32, 64} {
		b.Run(fmt.Sprintf("headers%d", count), func(b *testing.B) {
			headers := make(map[string][]string, count)
			for i := 0; i < count; i++ {
				headers[fmt.Sprintf("X-Bench-Header-%02d", i)] = []string{strings.Repeat("v", 64)}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkMetadata = New(headers)
			}
		})
	}
}
