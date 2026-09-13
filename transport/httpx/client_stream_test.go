package httpx

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

type streamWriter struct {
	bytes.Buffer
	closes   int
	maxWrite int
	err      error
	onWrite  func()
}

func (w *streamWriter) Close() error { w.closes++; return nil }
func (w *streamWriter) Write(p []byte) (int, error) {
	if w.maxWrite > 0 && len(p) > w.maxWrite {
		p = p[:w.maxWrite]
	}
	n, _ := w.Buffer.Write(p)
	if w.onWrite != nil {
		w.onWrite()
	}
	return n, w.err
}

func TestDownloadLimitsAndPartialWrites(t *testing.T) {
	for _, tc := range []struct {
		name         string
		body         string
		limit        int64
		length       int64
		maxWrite     int
		writeErr     error
		want         int64
		large, short bool
	}{
		{"unlimited", "abcdef", 0, -1, 0, nil, 6, false, false},
		{"exact", "abc", 3, -1, 0, nil, 3, false, false},
		{"excess", "abcd", 3, -1, 0, nil, 3, true, false},
		{"declared-excess", "abcd", 3, 4, 0, nil, 0, true, false},
		{"short-write", "abcd", 0, -1, 2, nil, 2, false, true},
		{"writer-error", "abcd", 0, -1, 2, io.ErrClosedPipe, 2, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{Reader: strings.NewReader(tc.body)}
			c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: 200, Header: http.Header{"X-File": {"yes"}}, Body: body, ContentLength: tc.length}, nil
			})))
			dst := &streamWriter{maxWrite: tc.maxWrite, err: tc.writeErr}
			old := 0
			info, err := c.Download(context.Background(), "/", dst, DownloadMaxBytes(tc.limit), OnResponse(func(*http.Response) error { old++; return nil }))
			var large *ResponseTooLargeError
			if info.StatusCode != 200 || info.Header.Get("X-File") != "yes" || info.Bytes != tc.want || int64(dst.Len()) != tc.want || errors.As(err, &large) != tc.large || errors.Is(err, io.ErrShortWrite) != tc.short || dst.closes != 0 || body.closes.Load() != 1 {
				t.Fatalf("info=%+v err=%v writer=%+v closes=%d", info, err, dst, body.closes.Load())
			}
			if tc.writeErr != nil && !errors.Is(err, tc.writeErr) {
				t.Fatal(err)
			}
			if (err == nil) != (old == 1) {
				t.Fatalf("completion hook called on failure: %d %v", old, err)
			}
		})
	}
}

func TestDownloadErrorResponsesAndHeaderHookFailure(t *testing.T) {
	for _, status := range []int{204, 205, 302, 400, 500} {
		body := &trackedBody{Reader: strings.NewReader(`{"message":"error page","code":123}`)}
		c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
		})))
		dst := &streamWriter{}
		info, err := c.Download(context.Background(), "/", dst)
		if info.StatusCode != status || info.Bytes != 0 || dst.Len() != 0 || body.closes.Load() != 1 || (err != nil) != (status >= 300) {
			t.Fatalf("status=%d info=%+v err=%v closes=%d", status, info, err, body.closes.Load())
		}
		if status >= 400 && errorx.FromError(err).Status != int32(status) {
			t.Fatal(err)
		}
	}
	body := &trackedBody{Reader: strings.NewReader("body")}
	c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: body, Header: http.Header{"X-Meta": {"yes"}}}, nil
	})))
	cause := errors.New("reject")
	dst := &streamWriter{}
	info, err := c.Download(context.Background(), "/", dst, OnResponseHeaders(func(context.Context, ResponseMeta) error { return cause }))
	if !errors.Is(err, cause) || info.Header.Get("X-Meta") != "yes" || info.Bytes != 0 || body.closes.Load() != 1 {
		t.Fatalf("%+v %v", info, err)
	}
}

func TestDownloadErrorDecodeLimit(t *testing.T) {
	decode, _ := NewResponseDecoder(WithDecodeMaxBytes(3))
	c, _ := NewClient(WithEndpoint("example.test"), WithResponseDecoder(decode), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader("too large error page"))}, nil
	})))
	var dst bytes.Buffer
	_, err := c.Download(context.Background(), "/", &dst, DownloadMaxBytes(1024))
	var large *ResponseTooLargeError
	if !errors.As(err, &large) || large.Limit != 3 || large.StatusCode != 400 || dst.Len() != 0 {
		t.Fatal(err)
	}
}

func TestInvokeReaderOwnershipAndRequestMetadata(t *testing.T) {
	body := &trackedBody{Reader: strings.NewReader("upload")}
	response := &trackedBody{Reader: strings.NewReader(`{"ok":true}`)}
	var order []string
	c, err := NewClient(WithEndpoint("http://example.test/base"), WithRequestEncoder(func(context.Context, string, interface{}) ([]byte, error) {
		t.Error("upload was encoded")
		return nil, nil
	}),
		WithMiddlewares(func(next middleware.ServiceHandler) middleware.ServiceHandler {
			return func(ctx context.Context, in interface{}) (interface{}, error) {
				order = append(order, "before")
				tr, _ := transport.FromContext(ctx)
				tr.In().Set("X-Auth", "yes")
				out, err := next(ctx, in)
				order = append(order, "after")
				return out, err
			}
		}), WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.URL.Path != "/base/upload" || r.Header.Get("X-Auth") != "yes" || r.Header.Get("Content-Type") != "application/octet-stream" {
				t.Errorf("request %s %v", r.URL, r.Header)
			}
			data, err := io.ReadAll(r.Body)
			if err != nil || string(data) != "upload" {
				t.Error(string(data), err)
			}
			_ = r.Body.Close() // Simulate Transport ownership; caller cleanup must not double-close.
			return &http.Response{StatusCode: 200, Body: response, Header: make(http.Header)}, nil
		})), WithResponseHeadersHook(func(context.Context, ResponseMeta) error { order = append(order, "headers"); return nil }))
	if err != nil {
		t.Fatal(err)
	}
	var reply struct {
		OK bool `json:"ok"`
	}
	err = c.InvokeReader(context.Background(), "POST", "/upload", body, &reply, SetRequestContentType("application/octet-stream"), OnResponse(func(*http.Response) error { order = append(order, "old"); return nil }))
	if err != nil || !reply.OK || body.closes.Load() != 1 || response.closes.Load() != 1 || strings.Join(order, ",") != "before,headers,old,after" {
		t.Fatalf("%v reply=%v order=%v close=%d/%d", err, reply, order, body.closes.Load(), response.closes.Load())
	}
}

func TestInvokeReaderClosesOnEarlyExits(t *testing.T) {
	for _, tc := range []struct {
		name, endpoint, path, method string
		ctx                          context.Context
		opts                         []CallOption
		middleware                   bool
	}{
		{"missing-endpoint", "", "/", "POST", context.Background(), nil, false},
		{"bad-path", "example.test", "bad", "POST", context.Background(), nil, false},
		{"bad-method", "example.test", "/", "bad method", context.Background(), nil, false},
		{"nil-context", "example.test", "/", "POST", nil, nil, false},
		{"bad-option", "example.test", "/", "POST", context.Background(), []CallOption{OnResponseHeaders(nil)}, false},
		{"middleware", "example.test", "/", "POST", context.Background(), nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := &trackedBody{Reader: strings.NewReader("upload")}
			opts := []ClientOption{WithEndpoint(tc.endpoint), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
				t.Error("sent after validation or middleware failure")
				return nil, errors.New("unexpected send")
			}))}
			if tc.middleware {
				opts = append(opts, WithMiddlewares(func(middleware.ServiceHandler) middleware.ServiceHandler {
					return func(context.Context, interface{}) (interface{}, error) { return nil, errors.New("stopped") }
				}))
			}
			c, _ := NewClient(opts...)
			if err := c.InvokeReader(tc.ctx, tc.method, tc.path, body, nil, tc.opts...); err == nil || body.closes.Load() != 1 {
				t.Fatalf("err=%v close=%d", err, body.closes.Load())
			}
		})
	}
}

func TestStreamMiddlewareCannotReplay(t *testing.T) {
	var requests atomic.Int32
	c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
		requests.Add(1)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("data"))}, nil
	})),
		WithMiddlewares(func(next middleware.ServiceHandler) middleware.ServiceHandler {
			return func(ctx context.Context, in interface{}) (interface{}, error) {
				if _, err := next(ctx, in); err != nil {
					return nil, err
				}
				return next(ctx, in)
			}
		}))
	var dst bytes.Buffer
	info, err := c.Download(context.Background(), "/", &dst)
	if err == nil || requests.Load() != 1 || info.Bytes != 4 || dst.String() != "data" {
		t.Fatal(info, err, dst.String())
	}
}

func TestInvokeReaderPreservesKnownReaderReplayMetadata(t *testing.T) {
	c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.ContentLength != 3 || r.GetBody == nil {
			t.Errorf("length=%d GetBody=%v", r.ContentLength, r.GetBody != nil)
		}
		return &http.Response{StatusCode: 204, Body: http.NoBody}, nil
	})))
	if err := c.InvokeReader(context.Background(), "POST", "/", strings.NewReader("abc"), nil); err != nil {
		t.Fatal(err)
	}
}

func TestStreamingHTTPIncrementalTransferAndCancellation(t *testing.T) {
	uploadFirst := make(chan struct{})
	downloadContinue := make(chan struct{})
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/upload":
			buf := make([]byte, 3)
			if _, err := io.ReadFull(r.Body, buf); err != nil {
				return
			}
			close(uploadFirst)
			if _, err := io.Copy(io.Discard, r.Body); err != nil {
				return
			}
			w.WriteHeader(204)
		case "/download":
			_, _ = io.WriteString(w, "one")
			w.(http.Flusher).Flush()
			select {
			case <-downloadContinue:
				_, _ = io.WriteString(w, "two")
			case <-r.Context().Done():
			}
		case "/cancel":
			_, _ = io.WriteString(w, "one")
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		}
	}))
	defer s.Close()
	c, _ := NewClient(WithEndpoint(s.URL), WithTimeout(0))
	defer c.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	producerDone := make(chan error, 1)
	go func() {
		if _, err := writer.Write([]byte("one")); err != nil {
			producerDone <- err
			return
		}
		select {
		case <-uploadFirst:
		case <-ctx.Done():
			producerDone <- ctx.Err()
			return
		}
		_, err := writer.Write([]byte("two"))
		writer.Close()
		producerDone <- err
	}()
	if err := c.InvokeReader(ctx, "POST", "/upload", reader, nil); err != nil {
		t.Fatal(err)
	}
	if err := <-producerDone; err != nil {
		t.Fatal(err)
	}
	var continued atomic.Bool
	dst := &streamWriter{onWrite: func() {
		if continued.CompareAndSwap(false, true) {
			close(downloadContinue)
		}
	}}
	info, err := c.Download(ctx, "/download", dst)
	if err != nil || info.Bytes != 6 || dst.String() != "onetwo" {
		t.Fatalf("download was buffered or failed: %v %v %q", info, err, dst.String())
	}
	cancelCtx, cancelRead := context.WithCancel(ctx)
	defer cancelRead()
	dst = &streamWriter{onWrite: cancelRead}
	info, err = c.Download(cancelCtx, "/cancel", dst)
	if !errors.Is(err, context.Canceled) || info.Bytes != 3 {
		t.Fatalf("cancel after partial download: %+v %v", info, err)
	}
}

func TestCopyDownloadReadFailureAndNoProgress(t *testing.T) {
	var dst bytes.Buffer
	n, err := copyDownload(context.Background(), &dst, io.MultiReader(strings.NewReader("abc"), errorReader{}), 0, 200)
	if n != 3 || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatal(n, err)
	}
	n, err = copyDownload(context.Background(), &dst, emptyReader{}, 0, 200)
	if n != 0 || !errors.Is(err, io.ErrNoProgress) {
		t.Fatal(n, err)
	}
}

type emptyReader struct{}

func (emptyReader) Read([]byte) (int, error) { return 0, nil }

type zeroStream struct{}

func (zeroStream) Read(p []byte) (int, error) { return len(p), nil }

func BenchmarkCopyDownload(b *testing.B) {
	for _, size := range []int64{1 << 20, 64 << 20} {
		name := "1MiB"
		if size > 1<<20 {
			name = "64MiB"
		}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(size)
			for i := 0; i < b.N; i++ {
				n, err := copyDownload(context.Background(), io.Discard, io.LimitReader(zeroStream{}, size), 0, 200)
				if err != nil || n != size {
					b.Fatal(n, err)
				}
			}
		})
	}
}

func TestDownloadGzipLimitAndCompletionHookError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Encoding", "gzip")
		zw := gzip.NewWriter(w)
		_, _ = io.WriteString(zw, strings.Repeat("a", 2048))
		_ = zw.Close()
	}))
	defer server.Close()
	c, _ := NewClient(WithEndpoint(server.URL))
	defer c.Close()
	var dst bytes.Buffer
	info, err := c.Download(context.Background(), "/", &dst, DownloadMaxBytes(1024))
	var large *ResponseTooLargeError
	if !errors.As(err, &large) || info.Bytes != 1024 || dst.Len() != 1024 {
		t.Fatalf("decompressed limit: %+v %v", info, err)
	}
	dst.Reset()
	cause := errors.New("completion hook")
	info, err = c.Download(context.Background(), "/", &dst, OnResponse(func(*http.Response) error { return cause }))
	if !errors.Is(err, cause) || info.Bytes != 2048 {
		t.Fatal(info, err)
	}
}

func TestStreamInputValidation(t *testing.T) {
	c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Error("invalid call sent")
		return nil, errors.New("unexpected send")
	})))
	if err := c.InvokeReader(context.Background(), "POST", "/", (*trackedBody)(nil), nil); err == nil {
		t.Fatal("nil reader accepted")
	}
	if _, err := c.Download(context.Background(), "/", (*streamWriter)(nil)); err == nil {
		t.Fatal("nil writer accepted")
	}
	if _, err := c.Download(context.Background(), "/", io.Discard, DownloadMaxBytes(-1)); err == nil {
		t.Fatal("negative limit accepted")
	}
}
