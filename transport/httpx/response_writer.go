package httpx

import (
	"bufio"
	"net"
	"net/http"

	"github.com/ml444/gkit/transport"
)

type transportResponseWriter struct {
	http.ResponseWriter
	tr      transport.ITransport
	flushed bool
}

func newTransportResponseWriter(w http.ResponseWriter, tr transport.ITransport) *transportResponseWriter {
	return &transportResponseWriter{
		ResponseWriter: w,
		tr:             tr,
	}
}

func (w *transportResponseWriter) WriteHeader(statusCode int) {
	w.flushTransportHeaders()
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *transportResponseWriter) Write(p []byte) (int, error) {
	w.flushTransportHeaders()
	return w.ResponseWriter.Write(p)
}

func (w *transportResponseWriter) flushTransportHeaders() {
	if w.flushed || w.tr == nil {
		return
	}
	w.flushed = true
	for k, v := range w.tr.Out() {
		if len(v) == 0 {
			continue
		}
		w.Header().Set(k, v[0])
	}
}

func (w *transportResponseWriter) Flush() {
	w.flushTransportHeaders()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (w *transportResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	w.flushTransportHeaders()
	return h.Hijack()
}

func (w *transportResponseWriter) Push(target string, opts *http.PushOptions) error {
	p, ok := w.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return p.Push(target, opts)
}
