package logging

import (
	"net/http"
	"time"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

// HTTPMiddleware logs structured HTTP access lines.
func HTTPMiddleware() middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			log.Infof("access method=%s path=%s status=%d bytes=%d latency_ms=%d trace=%s span=%s",
				r.Method, r.URL.Path, sw.status, sw.bytes, time.Since(start).Milliseconds(),
				header.LogTraceID(r), header.TraceInfoFromRequest(r).SpanID,
			)
		})
	}
}
