package grpcx

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	grpcmd "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/xds"

	"github.com/ml444/gkit/internal/netx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

type iServer interface {
	RegisterService(*grpc.ServiceDesc, any)
	Serve(net.Listener) error
	Stop()
	GracefulStop()
	GetServiceInfo() map[string]grpc.ServiceInfo
}

type Server struct {
	iServer
	name                string
	network             string
	address             string
	endpoint            string
	middlewares         []middleware.Middleware
	grpcOpts            []grpc.ServerOption
	unaryInterceptors   []grpc.UnaryServerInterceptor
	streamInterceptors  []grpc.StreamServerInterceptor
	credentials         credentials.TransportCredentials
	tlsConf             *tls.Config
	health              *health.Server
	timeout             time.Duration
	enableHealth        bool
	debug               bool
	disableTransportCtx bool
	enableXDS           bool
}

func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		name:    "gkit",
		network: "tcp",
		address: ":5040",
		health:  health.NewServer(),
	}
	s.unaryInterceptors = append(s.unaryInterceptors, s.defaultUnaryInterceptor())
	s.streamInterceptors = append(s.streamInterceptors, s.defaultStreamInterceptor())
	for _, o := range opts {
		o(s)
	}
	s.grpcOpts = append(s.grpcOpts,
		grpc.ChainUnaryInterceptor(s.unaryInterceptors...),
		grpc.ChainStreamInterceptor(s.streamInterceptors...),
	)
	if s.tlsConf != nil {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(credentials.NewTLS(s.tlsConf)))
	} else if s.credentials != nil {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(s.credentials))
	} else {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(insecure.NewCredentials()))
	}
	if s.enableXDS {
		var err error
		s.iServer, err = xds.NewGRPCServer(s.grpcOpts...)
		if err != nil {
			panic(err.Error())
		}
	} else {
		s.iServer = grpc.NewServer(s.grpcOpts...)
	}
	if s.enableHealth {
		s.health.SetServingStatus(s.name, healthpb.HealthCheckResponse_SERVING)
		healthpb.RegisterHealthServer(s.iServer, s.health)
	}
	if s.debug {
		reflection.Register(s.iServer)
	}
	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen(s.network, s.address)
	if err != nil {
		log.Errorf("net.Listen(%s, %s) failed: %v", s.network, s.address, err)
		return err
	}
	s.endpoint, _ = netx.ExtractEndpoint(s.address, ln)

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

func (s *Server) defaultUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if s.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, s.timeout)
			defer cancel()
		}
		if !s.disableTransportCtx {
			inMD, _ := grpcmd.FromIncomingContext(ctx)
			outMD := grpcmd.MD{}
			tr := &Transport{
				endpoint:  s.endpoint,
				operation: info.FullMethod,
				inMD:      transport.MD(inMD),
				outMD:     transport.MD(outMD),
			}
			ctx = transport.ToContext(ctx, tr)
			defer func() {
				if len(outMD) > 0 {
					_ = grpc.SetHeader(ctx, outMD)
				}
			}()
		}

		h := func(ctx context.Context, req interface{}) (interface{}, error) {
			return handler(ctx, req)
		}
		if len(s.middlewares) > 0 {
			h = middleware.Chain(s.middlewares...)(h)
		}
		return h(ctx, req)
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

// defaultStreamInterceptor is a gRPC stream server interceptor
func (s *Server) defaultStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !s.disableTransportCtx {
			ctx := ss.Context()
			inMD, _ := grpcmd.FromIncomingContext(ctx)
			outMD := grpcmd.MD{}
			ctx = transport.ToContext(ctx, &Transport{
				endpoint:  s.endpoint,
				operation: info.FullMethod,
				inMD:      transport.MD(inMD),
				outMD:     transport.MD(outMD),
			})
			ss = NewWrappedStream(ctx, ss)
			defer func() {
				if len(outMD) > 0 {
					_ = grpc.SetHeader(ctx, outMD)
				}
			}()
		}
		return handler(srv, ss)
	}
}
