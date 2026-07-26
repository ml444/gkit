package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/ml444/gkit/middleware"
)

// Options configures gzip compression.
type Options struct {
	MinLength int
	Level     int
}

type gzipResponseWriter struct {
	http.ResponseWriter
	writer    io.Writer
	gz        *gzip.Writer
	minLength int
	level     int
	buf       []byte
	enabled   bool
	pool      *sync.Pool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.enabled {
		w.buf = append(w.buf, b...)
		if len(w.buf) >= w.minLength {
			var err error
			w.gz = w.pool.Get().(*gzip.Writer)
			w.gz.Reset(w.ResponseWriter)

			w.writer = w.gz
			w.enabled = true
			if _, err = w.writer.Write(w.buf); err != nil {
				return 0, err
			}
			w.buf = nil
			return len(b), nil
		}
		return len(b), nil
	}
	return w.writer.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	if !w.enabled {
		_, err := w.ResponseWriter.Write(w.buf)
		return err
	}
	// 关闭后将 gzip.Writer 放回 pool 中复用
	err := w.gz.Close()
	// 重置为 io.Discard，防止 sync.Pool 长期持有 ResponseWriter 引用导致内存泄漏
	w.gz.Reset(io.Discard)
	w.pool.Put(w.gz)

	return err
}

// HTTPMiddleware compresses responses when client accepts gzip.
func HTTPMiddleware(opt Options) middleware.HttpMiddleware {
	if opt.MinLength <= 0 {
		opt.MinLength = 1024
	}
	level := opt.Level
	if level == 0 {
		level = gzip.DefaultCompression
	}
	// 为当前压缩级别创建一个 sync.Pool
	pool := &sync.Pool{
		New: func() any {
			// 初始化时传入 io.Discard，实际使用时会被 Reset 覆盖
			gz, _ := gzip.NewWriterLevel(io.Discard, level)
			return gz
		},
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			acceptsGzip := false
			for _, enc := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
				enc = strings.TrimSpace(enc)
				// 兼容标准 "gzip" 或带权重的情况如 "gzip;q=1.0"
				if enc == "gzip" || strings.HasPrefix(enc, "gzip;") {
					acceptsGzip = true
					break
				}
			}

			if !acceptsGzip {
				next.ServeHTTP(w, r)
				return
			}
			gw := &gzipResponseWriter{
				ResponseWriter: w,
				minLength:      opt.MinLength,
				level:          level,
				pool:           pool,
			}
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			next.ServeHTTP(gw, r)
			_ = gw.Close()
		})
	}
}
