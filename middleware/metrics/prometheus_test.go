//go:build prometheus

package metrics

import (
	"testing"
	"time"

	"github.com/ml444/gkit/pkg/header"
)

func TestPrometheusObserveDurationWithTrace(t *testing.T) {
	rec := prometheusRecorder{}
	ti := header.TraceInfo{
		TraceID: "4bf92f3577b34da6a3ce929d0e0e4736",
		SpanID:  "00f067aa0ba902b7",
	}
	rec.ObserveDurationWithTrace("GET", "/test-prom-exemplar", time.Millisecond, ti)
	rec.ObserveDurationWithTrace("GET", "/test-prom-plain", time.Millisecond, header.TraceInfo{})
}
