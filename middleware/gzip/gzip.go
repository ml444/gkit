package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

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
	buf       []byte
	enabled   bool
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	if !w.enabled {
		w.buf = append(w.buf, b...)
		if len(w.buf) >= w.minLength {
			var err error
			w.gz, err = gzip.NewWriterLevel(w.ResponseWriter, gzip.DefaultCompression)
			if err != nil {
				return 0, err
			}
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
	return w.gz.Close()
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
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}
			gw := &gzipResponseWriter{ResponseWriter: w, minLength: opt.MinLength}
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Del("Content-Length")
			next.ServeHTTP(gw, r)
			_ = gw.Close()
		})
	}
}
