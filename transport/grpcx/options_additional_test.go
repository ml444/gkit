package grpcx

import (
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/ml444/gkit/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestServerOptions(t *testing.T) {
	s := &Server{}
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	Debug(true)(s)
	Network("tcp4")(s)
	Address("127.0.0.1:0")(s)
	Name("svc")(s)
	EnableXDS()(s)
	EnableHealth()(s)
	Credentials(insecure.NewCredentials())(s)
	TLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})(s)
	Timeout(time.Second)(s)
	Middlewares(func(next middleware.ServiceHandler) middleware.ServiceHandler { return next })(s)
	SetMiddlewares(func(next middleware.ServiceHandler) middleware.ServiceHandler { return next })(s)
	UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	})(s)
	StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	})(s)
	Options(grpc.EmptyServerOption{})(s)
	Listener(lis)(s)
	DisableTransportCtx()(s)
	DisableErrorInterceptor()(s)

	if !s.debug || s.network != "tcp4" || s.address != "127.0.0.1:0" || s.name != "svc" ||
		!s.enableXDS || !s.enableHealth || s.credentials == nil || s.tlsConf == nil ||
		s.timeout != time.Second || len(s.middlewares) != 2 || len(s.unaryInterceptors) != 1 ||
		len(s.streamInterceptors) != 1 || len(s.grpcOpts) != 1 || s.listener != lis ||
		!s.disableTransportCtx || !s.disableErrorInterceptor {
		t.Fatalf("server options not applied: %#v", s)
	}
}

func TestClientOptions(t *testing.T) {
	c := &Client{}
	WithEndpoint("target")(c)
	WithDiscovery(nil, "svc")(c)
	WithTimeout(time.Second)(c)
	WithTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12})(c)
	WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(ctx, method, req, reply, cc, opts...)
	})(c)
	WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials()))(c)

	if c.endpoint != "target" || c.service != "svc" || c.timeout != time.Second ||
		c.tlsConf == nil || len(c.unaryInterceptors) != 1 || len(c.dialOpts) != 1 {
		t.Fatalf("client options not applied: %#v", c)
	}
}
