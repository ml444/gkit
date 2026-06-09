package httpx

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ml444/gkit/transport"
)

func TestTransportResponseWriterOptionalInterfaces(t *testing.T) {
	tr := &Transport{outMD: transport.Pairs("X-Test", "1")}
	rec := httptest.NewRecorder()
	tw := newTransportResponseWriter(rec, tr)
	tw.Flush()
	if rec.Header().Get("X-Test") != "1" {
		t.Fatalf("flush header = %#v", rec.Header())
	}
	if _, _, err := tw.Hijack(); err != http.ErrNotSupported {
		t.Fatalf("hijack err = %v", err)
	}
	if err := tw.Push("/x", nil); err != http.ErrNotSupported {
		t.Fatalf("push err = %v", err)
	}

	tw = newTransportResponseWriter(httptest.NewRecorder(), nil)
	tw.flushTransportHeaders()
}

func TestTransportResponseWriterDelegatesOptionalInterfaces(t *testing.T) {
	conn1, conn2 := net.Pipe()
	defer conn2.Close()
	tr := &Transport{outMD: transport.Pairs("X-Test", "1")}
	base := &optionalResponseWriter{
		header: http.Header{},
		conn:   conn1,
	}
	tw := newTransportResponseWriter(base, tr)
	conn, _, err := tw.Hijack()
	if err != nil {
		t.Fatalf("hijack: %v", err)
	}
	_ = conn.Close()
	if base.header.Get("X-Test") != "1" || !base.hijacked {
		t.Fatalf("hijack state header=%#v hijacked=%v", base.header, base.hijacked)
	}
	if err := tw.Push("/asset", nil); err != nil || base.pushed != "/asset" {
		t.Fatalf("push = %q %v", base.pushed, err)
	}
}

type optionalResponseWriter struct {
	header   http.Header
	conn     net.Conn
	hijacked bool
	pushed   string
}

func (w *optionalResponseWriter) Header() http.Header { return w.header }
func (w *optionalResponseWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
func (w *optionalResponseWriter) WriteHeader(int) {}
func (w *optionalResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return w.conn, bufio.NewReadWriter(bufio.NewReader(w.conn), bufio.NewWriter(w.conn)), nil
}
func (w *optionalResponseWriter) Push(target string, _ *http.PushOptions) error {
	w.pushed = target
	return nil
}
