package grpcx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	grpcmd "google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"

	"github.com/ml444/gkit/internal/netx"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/middleware/general"
	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/grpcx/xds"
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
	listener            net.Listener
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
	disableErrorInterceptor bool
	enableXDS               bool
}

// NewServer creates a gRPC server.
func NewServer(opts ...ServerOption) (*Server, error) {
	s := &Server{
		name:    "gkit",
		network: "tcp",
		address: ":5040",
	}
	s.unaryInterceptors = append(s.unaryInterceptors, s.defaultUnaryInterceptor())
	s.streamInterceptors = append(s.streamInterceptors, s.defaultStreamInterceptor())
	for _, o := range opts {
		o(s)
	}
	if !s.disableErrorInterceptor {
		s.unaryInterceptors = append(s.unaryInterceptors, general.ServerErrorInterceptor)
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
		gs, err := xds.NewGRPCServer(s.grpcOpts...)
		if err != nil {
			return nil, err
		}
		s.iServer = gs
	} else {
		s.iServer = grpc.NewServer(s.grpcOpts...)
	}
	if s.enableHealth {
		s.health = health.NewServer()
		s.health.SetServingStatus(s.name, healthpb.HealthCheckResponse_SERVING)
		healthpb.RegisterHealthServer(s.iServer, s.health)
	}
	if s.debug {
		if gs, ok := s.iServer.(*grpc.Server); ok {
			reflection.Register(gs)
		}
	}
	return s, nil
}

// Endpoint returns the listen address in host:port form.
func (s *Server) Endpoint() (string, error) {
	if err := s.listenAndEndpoint(); err != nil {
		return "", err
	}
	return s.endpoint, nil
}

func (s *Server) listenAndEndpoint() error {
	if s.listener == nil {
		ln, err := net.Listen(s.network, s.address)
		if err != nil {
			return err
		}
		s.listener = ln
	}
	if s.endpoint == "" {
		addr, err := netx.ExtractEndpoint(s.address, s.listener)
		if err != nil {
			return err
		}
		if addr == "" {
			return errors.New("grpcx: failed to extract endpoint")
		}
		s.endpoint = addr
	}
	return nil
}

// Start starts the gRPC server and blocks until it stops.
func (s *Server) Start() error {
	if err := s.listenAndEndpoint(); err != nil {
		log.Errorf("grpcx: listen endpoint failed: %v", err)
		return err
	}
	log.Infof("[gRPC] server listening on: %s", s.listener.Addr().String())
	if s.enableHealth && s.health != nil {
		s.health.Resume()
	}
	return s.Serve(s.listener)
}

// Stop gracefully stops the server. If ctx is canceled, forces Stop.
func (s *Server) Stop(ctx context.Context) error {
	if s.enableHealth && s.health != nil {
		s.health.Shutdown()
	}
	done := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		log.Info("[gRPC] server stopping")
		return nil
	case <-ctx.Done():
		s.iServer.Stop()
		log.Info("[gRPC] server stopping (forced)")
		return ctx.Err()
	}
}

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

type wrappedStream struct {
	grpc.ServerStream
	ctx        context.Context
	outMD      grpcmd.MD
	headerSent bool
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func (w *wrappedStream) SendMsg(m interface{}) error {
	if err := w.sendOutboundHeaders(); err != nil {
		return err
	}
	return w.ServerStream.SendMsg(m)
}

func (w *wrappedStream) SendHeader(md grpcmd.MD) error {
	if err := grpc.SendHeader(w.Context(), md); err != nil {
		return err
	}
	w.headerSent = true
	return nil
}

func (w *wrappedStream) sendOutboundHeaders() error {
	if w.headerSent || len(w.outMD) == 0 {
		return nil
	}
	if err := w.ServerStream.SetHeader(w.outMD); err != nil {
		return err
	}
	w.headerSent = true
	return nil
}

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
			ws := &wrappedStream{
				ServerStream: ss,
				ctx:          ctx,
				outMD:        outMD,
			}
			err := handler(srv, ws)
			_ = ws.sendOutboundHeaders()
			return err
		}
		return handler(srv, ss)
	}
}
