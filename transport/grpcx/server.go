package grpcx

import (
	"crypto/tls"
	"net"

	"github.com/ml444/gutil/netx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/xds"

	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/transport"
)

type Server struct {
	*xds.GRPCServer
	name         string
	network      string
	address      string
	creds        credentials.TransportCredentials
	tlsConf      *tls.Config
	grpcOpts     []grpc.ServerOption
	health       *health.Server
	customHealth bool
	debug        bool
}

func NewServer(registerFunc func(s grpc.ServiceRegistrar), opts ...ServerOption) *Server {
	s := &Server{
		name:    "gkit",
		network: "tcp",
		address: ":5040",
		health:  health.NewServer(),
	}
	for _, o := range opts {
		o(s)
	}
	if s.tlsConf != nil {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(credentials.NewTLS(s.tlsConf)))
	} else if s.creds != nil {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(s.creds))
	} else {
		s.grpcOpts = append(s.grpcOpts, grpc.Creds(insecure.NewCredentials()))
	}
	s.GRPCServer, _ = xds.NewGRPCServer(s.grpcOpts...)
	if !s.customHealth {
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
	transport.GrpcHostAddress, _ = netx.Extract(s.address, ln)

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
