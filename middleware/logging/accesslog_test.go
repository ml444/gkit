package logging

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware/tracing"
	"github.com/ml444/gkit/pkg/header"
)

type captureWriter struct{ buf bytes.Buffer }

func (w *captureWriter) Write(p []byte) (int, error)       { return w.buf.Write(p) }
func (w *captureWriter) WriteString(s string) (int, error) { return w.buf.WriteString(s) }

func TestHTTPMiddlewareAccessLogTraceparent(t *testing.T) {
	cw := &captureWriter{}
	log.SetLogger(log.NewDefaultLogger(cw))

	chain := tracing.HTTPMiddleware()(HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set(header.TraceparentHeaderKey, "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	rr := httptest.NewRecorder()
	chain.ServeHTTP(rr, req)

	out := cw.buf.String()
	if !strings.Contains(out, "trace=4bf92f3577b34da6a3ce929d0e0e4736") {
		t.Fatalf("access log missing trace: %s", out)
	}
	if !strings.Contains(out, "span=00f067aa0ba902b7") {
		t.Fatalf("access log missing span: %s", out)
	}
}
