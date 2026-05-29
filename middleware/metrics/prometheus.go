//go:build prometheus

package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	reqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "gkit_http_requests_total", Help: "Total HTTP requests"},
		[]string{"method", "path", "status"},
	)
	reqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "gkit_http_request_duration_seconds", Help: "HTTP request latency"},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(reqTotal, reqDuration)
	SetRecorder(prometheusRecorder{})
}

type prometheusRecorder struct{}

func (prometheusRecorder) IncRequests(method, path string, status int) {
	reqTotal.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
}

func (prometheusRecorder) ObserveDuration(method, path string, d time.Duration) {
	reqDuration.WithLabelValues(method, path).Observe(d.Seconds())
}
