package cryptox

import (
	"testing"
)

func BenchmarkMd5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Md5("hello")
	}
}
