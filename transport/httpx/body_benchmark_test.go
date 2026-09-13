package httpx

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ml444/gkit/log"
)

// BenchmarkBodyBinding includes routing, the default server middleware, reading,
// JSON decoding and a fixed response, but no network I/O or response encoding.
func BenchmarkBodyBinding(b *testing.B) {
	previous := log.GetLogger()
	log.SetLogger(log.NewNopLogger())
	b.Cleanup(func() { log.SetLogger(previous) })
	for _, size := range []int{128, 512, 1 << 10, 64 << 10, 1 << 20} {
		b.Run(fmt.Sprintf("bytes%d", size), func(b *testing.B) {
			body := `{"value":"` + strings.Repeat("x", size-len(`{"value":""}`)) + `"}`
			srv := NewServer()
			response := []byte("ok")
			srv.GetRouter().POST("/bind", func(w http.ResponseWriter, r *http.Request) {
				var v struct{ Value string }
				if err := NewCtx(w, r).Bind(&v); err != nil {
					b.Fatal(err)
				}
				_, _ = w.Write(response)
			})
			r := httptest.NewRequest(http.MethodPost, "/bind", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			var reader strings.Reader
			input := io.NopCloser(&reader)
			w := &bodyBenchmarkWriter{header: make(http.Header)}
			b.ReportAllocs()
			b.SetBytes(int64(size))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reader.Reset(body)
				r.Body = input
				srv.ServeHTTP(w, r)
			}
		})
	}
}

type bodyBenchmarkWriter struct{ header http.Header }

func (w *bodyBenchmarkWriter) Header() http.Header       { return w.header }
func (*bodyBenchmarkWriter) WriteHeader(int)             {}
func (*bodyBenchmarkWriter) Write(p []byte) (int, error) { return len(p), nil }
