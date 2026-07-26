package logging

import (
	"math/rand/v2"
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

type accessLogOptions struct {
	sampleRate    float64       // 采样率 (0.0 - 1.0)
	slowThreshold time.Duration // 慢请求阈值
}

type AccessLogOption func(*accessLogOptions)

// WithSampleRate 设置访问日志的采样率 (例如 0.1 表示 10% 的请求记录日志)
func WithSampleRate(rate float64) AccessLogOption {
	return func(o *accessLogOptions) {
		o.sampleRate = rate
	}
}

// WithSlowThreshold 设置慢请求阈值。当请求耗时超过此值时，必然打印日志（无视采样率），并可提升日志级别为 Warn
func WithSlowThreshold(threshold time.Duration) AccessLogOption {
	return func(o *accessLogOptions) {
		o.slowThreshold = threshold
	}
}

// HTTPMiddleware logs structured HTTP access lines.
func HTTPMiddleware(opts ...AccessLogOption) middleware.HttpMiddleware {
	// 默认配置：100% 采样，不开启慢请求阈值
	options := &accessLogOptions{
		sampleRate:    1.0,
		slowThreshold: 0,
	}
	for _, opt := range opts {
		opt(options)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			latency := time.Since(start)
			latencyMs := latency.Milliseconds()

			// 1. 慢请求判断
			isSlow := options.slowThreshold > 0 && latency >= options.slowThreshold

			// 2. 采样逻辑拦截 (如果不是慢请求，且未命中采样，则直接丢弃日志)
			if !isSlow && options.sampleRate < 1.0 {
				if options.sampleRate <= 0.0 || rand.Float64() > options.sampleRate {
					return
				}
			}

			// 3. 日志打印
			format := "access method=%s path=%s status=%d bytes=%d latency_ms=%d trace=%s span=%s"
			args := []any{
				r.Method, r.URL.Path, sw.status, sw.bytes, latencyMs,
				header.LogTraceID(r), header.TraceInfoFromRequest(r).SpanID,
			}

			if isSlow {
				// 慢请求通常需要重点关注，使用 Warnf 级别，并打上特殊标记
				log.Warnf("[SLOW_REQ] "+format, args...)
			} else {
				// 普通请求命中采样后，正常使用 Infof
				log.Infof(format, args...)
			}
		})
	}
}
