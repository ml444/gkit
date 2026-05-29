package metrics

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

// Recorder records request metrics.
type Recorder interface {
	IncRequests(method, path string, status int)
	ObserveDuration(method, path string, duration time.Duration)
}

type nopRecorder struct{}

func (nopRecorder) IncRequests(string, string, int)              {}
func (nopRecorder) ObserveDuration(string, string, time.Duration) {}

var defaultRecorder Recorder = nopRecorder{}

// SetRecorder sets the global metrics recorder (e.g. Prometheus).
func SetRecorder(r Recorder) {
	if r != nil {
		defaultRecorder = r
	}
}

// HTTPMiddleware records HTTP request metrics.
func HTTPMiddleware() middleware.HttpMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusCapture{ResponseWriter: w}
			next.ServeHTTP(sw, r)
			defaultRecorder.IncRequests(r.Method, r.URL.Path, sw.status)
			defaultRecorder.ObserveDuration(r.Method, r.URL.Path, time.Since(start))
		})
	}
}

// Server records service-layer metrics using transport path.
func Server() middleware.Middleware {
	return func(next middleware.ServiceHandler) middleware.ServiceHandler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			start := time.Now()
			path := "unknown"
			if tr, ok := transport.FromContext(ctx); ok {
				path = tr.Path()
			}
			rsp, err := next(ctx, req)
			status := 200
			if err != nil {
				status = 500
			}
			defaultRecorder.IncRequests("RPC", path, status)
			defaultRecorder.ObserveDuration("RPC", path, time.Since(start))
			return rsp, err
		}
	}
}

type statusCapture struct {
	http.ResponseWriter
	status int
}

func (w *statusCapture) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusCapture) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

// InMemoryRecorder is a simple thread-safe recorder for tests.
type InMemoryRecorder struct {
	mu        sync.Mutex
	Requests  map[string]int
	Durations []time.Duration
}

func NewInMemoryRecorder() *InMemoryRecorder {
	return &InMemoryRecorder{Requests: make(map[string]int)}
}

func (r *InMemoryRecorder) IncRequests(method, path string, status int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := method + " " + path + " " + http.StatusText(status)
	r.Requests[key]++
}

func (r *InMemoryRecorder) ObserveDuration(method, path string, d time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Durations = append(r.Durations, d)
}
