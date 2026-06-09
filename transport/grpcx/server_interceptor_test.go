package grpcx

import (
	"context"
	"testing"

	"github.com/ml444/gkit/transport"
	"google.golang.org/grpc"
	grpcmd "google.golang.org/grpc/metadata"
)

func TestDefaultUnaryInterceptorTransportContext(t *testing.T) {
	s := &Server{endpoint: "127.0.0.1:1"}
	interceptor := s.defaultUnaryInterceptor()
	called := false
	_, err := interceptor(grpcmd.NewIncomingContext(context.Background(), grpcmd.Pairs("in", "1")), "req",
		&grpc.UnaryServerInfo{FullMethod: "/svc/Method"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			tr, ok := transport.FromContext(ctx)
			if !ok {
				t.Fatal("missing transport context")
			}
			if tr.Kind() != "grpc" || tr.Endpoint() != "127.0.0.1:1" || tr.Path() != "/svc/Method" || tr.In().GetFirst("in") != "1" {
				t.Fatalf("unexpected transport: %#v", tr)
			}
			tr.Out().Set("out", "2")
			return "resp", nil
		})
	if err != nil || !called {
		t.Fatalf("interceptor err=%v called=%v", err, called)
	}
}

func TestDefaultUnaryInterceptorDisabledTransport(t *testing.T) {
	s := &Server{disableTransportCtx: true}
	interceptor := s.defaultUnaryInterceptor()
	_, err := interceptor(context.Background(), "req", &grpc.UnaryServerInfo{FullMethod: "/svc/Method"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			if _, ok := transport.FromContext(ctx); ok {
				t.Fatal("unexpected transport context")
			}
			return nil, nil
		})
	if err != nil {
		t.Fatal(err)
	}
}

func TestWrappedStreamAndDefaultStreamInterceptor(t *testing.T) {
	base := &fakeServerStream{ctx: grpcmd.NewIncomingContext(context.Background(), grpcmd.Pairs("in", "1"))}
	s := &Server{endpoint: "127.0.0.1:1"}
	interceptor := s.defaultStreamInterceptor()
	err := interceptor(nil, base, &grpc.StreamServerInfo{FullMethod: "/svc/Stream"}, func(_ interface{}, ss grpc.ServerStream) error {
		tr, ok := transport.FromContext(ss.Context())
		if !ok {
			t.Fatal("missing transport context")
		}
		tr.Out().Set("out", "2")
		return ss.SendMsg("hello")
	})
	if err != nil {
		t.Fatal(err)
	}
	if base.sent != 1 || base.header.Get("out")[0] != "2" {
		t.Fatalf("stream sent=%d header=%#v", base.sent, base.header)
	}

	ws := &wrappedStream{ServerStream: base, ctx: context.WithValue(context.Background(), "k", "v")}
	if ws.Context().Value("k") != "v" {
		t.Fatal("wrapped context mismatch")
	}
	if err := ws.SendHeader(grpcmd.Pairs("x", "y")); err == nil {
		t.Fatal("expected SendHeader error outside real grpc stream")
	}
}

type fakeServerStream struct {
	ctx    context.Context
	header grpcmd.MD
	sent   int
}

func (f *fakeServerStream) SetHeader(md grpcmd.MD) error {
	if f.header == nil {
		f.header = grpcmd.MD{}
	}
	for k, v := range md {
		f.header[k] = append(f.header[k], v...)
	}
	return nil
}

func (f *fakeServerStream) SendHeader(md grpcmd.MD) error {
	return f.SetHeader(md)
}

func (f *fakeServerStream) SetTrailer(grpcmd.MD) {}

func (f *fakeServerStream) Context() context.Context {
	if f.ctx == nil {
		return context.Background()
	}
	return f.ctx
}

func (f *fakeServerStream) SendMsg(interface{}) error {
	f.sent++
	return nil
}

func (f *fakeServerStream) RecvMsg(interface{}) error {
	return nil
}
