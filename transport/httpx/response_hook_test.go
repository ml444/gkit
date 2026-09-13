package httpx

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/errorx"
	"github.com/ml444/gkit/middleware"
)

type trackedBody struct {
	io.Reader
	closes atomic.Int32
}

func (b *trackedBody) Close() error { b.closes.Add(1); return nil }

func TestResponseHeadersHooksOrderAndIsolation(t *testing.T) {
	for _, status := range []int{200, 204, 400, 500} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			var order []string
			body := &trackedBody{Reader: strings.NewReader(`{"message":"remote"}`)}
			c, err := NewClient(WithEndpoint("http://example.test"), WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: status, Header: http.Header{"X-Test": {"original"}}, Body: body, ContentLength: -1, Request: r}, nil
			})), WithResponseHeadersHook(func(_ context.Context, m ResponseMeta) error {
				order = append(order, "client")
				m.Header.Set("X-Test", "changed")
				return nil
			}), WithResponseDecoder(func(_ context.Context, r *http.Response, _ interface{}) error {
				order = append(order, "decode")
				if r.Header.Get("X-Test") != "original" {
					t.Error("header mutated")
				}
				_, err := io.ReadAll(r.Body)
				if err != nil {
					return err
				}
				if status >= 400 {
					return errorx.CreateError(int32(status), 1, "remote")
				}
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			err = c.Invoke(context.Background(), "GET", "/", nil, nil, OnResponseHeaders(func(_ context.Context, m ResponseMeta) error {
				order = append(order, "call")
				if m.StatusCode != status || m.Header.Get("X-Test") != "original" {
					t.Error(m)
				}
				return nil
			}), OnResponse(func(*http.Response) error { order = append(order, "old"); return nil }))
			want := []string{"client", "call", "decode"}
			if status < 400 {
				want = append(want, "old")
			}
			if !reflect.DeepEqual(order, want) || (err != nil) != (status >= 400) || body.closes.Load() != 1 {
				t.Fatalf("%v %v closes=%d", order, err, body.closes.Load())
			}
		})
	}
}

func TestResponseHeadersHookFailureAndConfiguration(t *testing.T) {
	cause := errors.New("hook stopped")
	body := &trackedBody{Reader: strings.NewReader("unread")}
	var sends atomic.Int32
	c, err := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
		sends.Add(1)
		return &http.Response{StatusCode: 500, Header: make(http.Header), Body: body}, nil
	})),
		WithResponseDecoder(func(context.Context, *http.Response, interface{}) error {
			t.Error("decoder ran after rejected hook")
			return nil
		}))
	if err != nil {
		t.Fatal(err)
	}
	err = c.Invoke(context.Background(), "GET", "/", nil, nil, OnResponseHeaders(func(context.Context, ResponseMeta) error { return cause }), OnResponseHeaders(func(context.Context, ResponseMeta) error { t.Error("second hook ran"); return nil }))
	var hookErr *ResponseHookError
	if !errors.Is(err, cause) || !errors.As(err, &hookErr) || hookErr.StatusCode != 500 || body.closes.Load() != 1 {
		t.Fatalf("%v closes=%d", err, body.closes.Load())
	}
	if _, err := NewClient(WithResponseHeadersHook(nil)); err == nil {
		t.Fatal("nil client hook accepted")
	}
	if err := c.Invoke(context.Background(), "GET", "/", nil, nil, OnResponseHeaders(nil)); err == nil || sends.Load() != 1 {
		t.Fatal("nil call hook was sent")
	}
	if err := c.Invoke(context.Background(), "GET", "/", nil, nil, nil); err == nil || sends.Load() != 1 {
		t.Fatal("nil option was sent")
	}
}

func TestResponseHooksTransportAndDecodeFailures(t *testing.T) {
	for _, networkFailure := range []bool{false, true} {
		var hooks, old int
		cause := errors.New("network or decoder")
		c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if networkFailure {
				return nil, cause
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("bad JSON"))}, nil
		})), WithResponseHeadersHook(func(context.Context, ResponseMeta) error { hooks++; return nil }))
		err := c.Invoke(context.Background(), "GET", "/", nil, &struct{}{}, OnResponse(func(*http.Response) error { old++; return nil }))
		if err == nil || old != 0 || (networkFailure && hooks != 0) || (!networkFailure && hooks != 1) {
			t.Fatalf("hooks=%d old=%d err=%v", hooks, old, err)
		}
	}
}

func TestResponseHooksConcurrentIsolation(t *testing.T) {
	header := http.Header{"X-Value": {"original"}}
	c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 204, Header: header, Body: http.NoBody}, nil
	})),
		WithResponseHeadersHook(func(_ context.Context, m ResponseMeta) error { m.Header.Set("X-Value", "changed"); return nil }))
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Invoke(context.Background(), "GET", "/", nil, nil, OnResponseHeaders(func(_ context.Context, m ResponseMeta) error {
				if m.Header.Get("X-Value") != "original" {
					t.Error(m.Header)
				}
				return nil
			})); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
}

func TestResponseFeedbackClassification(t *testing.T) {
	local := errors.New("local")
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	for _, tc := range []struct {
		ctx                     context.Context
		status                  int
		readErr, err            error
		remote, success, report bool
	}{
		{context.Background(), 200, nil, nil, false, true, true},
		{context.Background(), 400, nil, local, true, true, true},
		{context.Background(), 200, nil, local, false, false, false},
		{context.Background(), 400, nil, local, false, false, false},
		{context.Background(), 500, nil, local, false, false, true},
		{context.Background(), 200, io.ErrUnexpectedEOF, local, false, false, true},
		{canceled, 200, context.Canceled, local, false, false, false},
	} {
		if success, report := responseFeedback(tc.ctx, tc.status, tc.readErr, tc.err, tc.remote); success != tc.success || report != tc.report {
			t.Fatalf("feedback=%v,%v want=%v,%v", success, report, tc.success, tc.report)
		}
	}
}

type recordingBalancer struct{ updates []bool }

func (b *recordingBalancer) Select(_ context.Context, instances []discovery.ServiceInstancer) (discovery.ServiceInstancer, error) {
	if len(instances) == 0 {
		return nil, discovery.ErrNotFound
	}
	return instances[0], nil
}
func (b *recordingBalancer) Update(_ context.Context, _ discovery.ServiceInstancer, success bool) {
	b.updates = append(b.updates, success)
}

func TestManagedCallDiscoveryFeedback(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		hookError bool
		want      []bool
	}{
		{"business-error", 400, `{"message":"bad request"}`, false, []bool{true}},
		{"decode-error", 200, `invalid`, false, nil},
		{"hook-error", 200, `{}`, true, nil},
		{"server-and-hook-error", 500, `{}`, true, []bool{false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			registry := discovery.NewDefaultRegistry()
			if err := registry.Register(context.Background(), &discovery.ServiceInstance{ID: "test", Name: "svc", Address: "127.0.0.1", Port: 8080}); err != nil {
				t.Fatal(err)
			}
			lb := &recordingBalancer{}
			dc := discovery.NewDiscoveryClient(registry, discovery.WithLoadBalancer(lb))
			c, err := NewClient(WithDiscovery(dc, "svc"), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: tc.status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(tc.body))}, nil
			})))
			if err != nil {
				t.Fatal(err)
			}
			var opts []CallOption
			if tc.hookError {
				opts = append(opts, OnResponseHeaders(func(context.Context, ResponseMeta) error { return errors.New("local hook") }))
			}
			if err := c.Invoke(context.Background(), "GET", "/", nil, &struct{}{}, opts...); err == nil {
				t.Fatal("expected error")
			}
			if !reflect.DeepEqual(lb.updates, tc.want) {
				t.Fatalf("feedback=%v want=%v", lb.updates, tc.want)
			}
		})
	}
}

func TestRawDoRetainsCallerBodyOwnershipAndSkipsHooks(t *testing.T) {
	body := &trackedBody{Reader: strings.NewReader("raw")}
	c, _ := NewClient(WithResponseHeadersHook(func(context.Context, ResponseMeta) error { t.Error("raw Do executed managed hook"); return nil }), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) { return &http.Response{StatusCode: 200, Body: body}, nil })))
	req, _ := http.NewRequest("GET", "http://example.test/", nil)
	r, err := c.Do(req)
	if err != nil || body.closes.Load() != 0 {
		t.Fatal(err, body.closes.Load())
	}
	_ = r.Body.Close()
	if body.closes.Load() != 1 {
		t.Fatal(body.closes.Load())
	}
}

func TestInvokePreservesNilMiddlewareResultOnErrors(t *testing.T) {
	for _, decodeFails := range []bool{true, false} {
		cause := errors.New("failure")
		c, _ := NewClient(WithEndpoint("example.test"), WithTransport(roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
		})),
			WithResponseDecoder(func(context.Context, *http.Response, interface{}) error {
				if decodeFails {
					return cause
				}
				return nil
			}),
			WithMiddlewares(func(next middleware.ServiceHandler) middleware.ServiceHandler {
				return func(ctx context.Context, in interface{}) (interface{}, error) {
					out, err := next(ctx, in)
					if out != nil || !errors.Is(err, cause) {
						t.Errorf("middleware result=%v err=%v", out, err)
					}
					return out, err
				}
			}))
		err := c.Invoke(context.Background(), "GET", "/", nil, &struct{}{}, OnResponse(func(*http.Response) error { return cause }))
		if !errors.Is(err, cause) {
			t.Fatal(err)
		}
	}
}
