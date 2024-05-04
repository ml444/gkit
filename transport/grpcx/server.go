package grpcx

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/ml444/gutil/netx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	grpcmd "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/xds"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/pkg/header"
	"github.com/ml444/gkit/transport"
)

type Server struct {
	*xds.GRPCServer
	name               string
	network            string
	address            string
	endpoint           string
	middlewares        []middleware.Middleware
	grpcOpts           []grpc.ServerOption
	unaryInterceptors  []grpc.UnaryServerInterceptor
	streamInterceptors []grpc.StreamServerInterceptor
	creds              credentials.TransportCredentials
	tlsConf            *tls.Config
	health             *health.Server
	timeout            time.Duration
	enableHealth       bool
	debug              bool
}

func NewServer(registerFunc func(s grpc.ServiceRegistrar), opts ...ServerOption) *Server {
	s := &Server{
		name:    "gkit",
		network: "tcp",
		address: ":5040",
		health:  health.NewServer(),
	}
	s.unaryInterceptors = append(s.unaryInterceptors, s.unaryServerInterceptor())
	s.streamInterceptors = append(s.streamInterceptors, s.streamServerInterceptor())
	for _, o := range opts {
		o(s)
	}
	s.grpcOpts = append(s.grpcOpts,
		grpc.ChainUnaryInterceptor(s.unaryInterceptors...),
		grpc.ChainStreamInterceptor(s.streamInterceptors...),
	)
	if s.tlsConf != nil {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(credentials.NewTLS(s.tlsConf)))
	} else if s.creds != nil {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(s.creds))
	} else {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(insecure.NewCredentials()))
	}
	s.GRPCServer, _ = xds.NewGRPCServer(s.grpcOpts...)
	if s.enableHealth {
		s.health.SetServingStatus(s.name, healthpb.HealthCheckResponse_SERVING)
		healthpb.RegisterHealthServer(s.GRPCServer, s.health)
	}
	if s.debug {
		reflection.Register(s.GRPCServer)
	}
	registerFunc(s.GRPCServer)
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen(s.network, s.address)
	if err != nil {
		log.Errorf("net.Listen(%s, %s) failed: %v", s.network, s.address, err)
		return err
	}
	s.endpoint, _ = netx.Extract(s.address, ln)

	log.Infof("[gRPC] server listening on: %s", ln.Addr().String())
	s.health.Resume()
	return s.Serve(ln)
}

func (s *Server) Stop() error {
	s.health.Shutdown()
	s.GracefulStop()
	log.Info("[gRPC] server stopping")
	return nil
}

/*
>>>>>>>>>>>> server interceptor <<<<<<<<<<<<<
*/

func (s *Server) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, _ := grpcmd.FromIncomingContext(ctx)
		outHeader := grpcmd.MD{}
		tr := &Transport{
			BaseTransport: transport.BaseTransport{
				Endpoint:  s.endpoint,
				Operation: info.FullMethod,
				InHeader:  header.Header(md),
				OutHeader: header.Header(outHeader),
			},
		}
		ctx = transport.ToContext(ctx, tr)
		if s.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, s.timeout)
			defer cancel()
		}
		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}
		if len(s.middlewares) > 0 {
			h = middleware.Chain(s.middlewares...)(h)
		}
		reply, err := h(ctx, req)
		if len(outHeader) > 0 {
			_ = grpc.SetHeader(ctx, outHeader)
		}
		return reply, err
	}
}

// wrappedStream is rewrite grpc stream's context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func NewWrappedStream(ctx context.Context, stream grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{
		ServerStream: stream,
		ctx:          ctx,
	}
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

// streamServerInterceptor is a gRPC stream server interceptor
func (s *Server) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		md, _ := grpcmd.FromIncomingContext(ctx)
		outHeader := grpcmd.MD{}
		ctx = transport.ToContext(ctx, &Transport{
			BaseTransport: transport.BaseTransport{
				Endpoint:  s.endpoint,
				Operation: info.FullMethod,
				InHeader:  header.Header(md),
				OutHeader: header.Header(outHeader),
			},
		})

		ws := NewWrappedStream(ctx, ss)

		err := handler(srv, ws)
		if len(outHeader) > 0 {
			_ = grpc.SetHeader(ctx, outHeader)
		}
		return err
	}
}
