//go:build !prometheus

package metrics

// Prometheus integration is optional. Build with -tags prometheus and add:
//
//	go get github.com/prometheus/client_golang/prometheus
//
// Then metrics.SetRecorder is wired automatically via prometheus.go init.
