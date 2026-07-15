//go:build prometheus

package metrics

import (
	"strconv"
	"time"

	"github.com/ml444/gkit/pkg/header"
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

func (prometheusRecorder) ObserveDurationWithTrace(method, path string, d time.Duration, ti header.TraceInfo) {
	observer := reqDuration.WithLabelValues(method, path)
	if ti.TraceID == "" {
		observer.Observe(d.Seconds())
		return
	}
	labels := prometheus.Labels{"trace_id": ti.TraceID}
	if ti.SpanID != "" {
		labels["span_id"] = ti.SpanID
	}
	if ex, ok := observer.(prometheus.ExemplarObserver); ok {
		ex.ObserveWithExemplar(d.Seconds(), labels)
		return
	}
	observer.Observe(d.Seconds())
}
