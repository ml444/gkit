package grpcx

import (
	"crypto/tls"
	"net"
	"time"

	"github.com/ml444/gkit/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServerOption is gRPC server option.
type ServerOption func(o *Server)

func Debug(debug bool) ServerOption {
	return func(s *Server) {
		s.debug = debug
	}
}

// Network with server network.
func Network(network string) ServerOption {
	return func(s *Server) {
		s.network = network
	}
}

// Address with server address.
func Address(addr string) ServerOption {
	return func(s *Server) {
		s.address = addr
	}
}

// Name with server name.
func Name(name string) ServerOption {
	return func(s *Server) {
		s.name = name
	}
}

// EnableXDS uses an xDS control-plane-backed gRPC server (requires GRPC_XDS_BOOTSTRAP).
func EnableXDS() ServerOption {
	return func(s *Server) {
		s.enableXDS = true
	}
}

// EnableHealth registers the gRPC health service.
func EnableHealth() ServerOption {
	return func(s *Server) {
		s.enableHealth = true
	}
}

// Credentials with server credentials.
func Credentials(creds credentials.TransportCredentials) ServerOption {
	return func(s *Server) {
		s.credentials = creds
	}
}

// TLSConfig with TLS config.
func TLSConfig(c *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConf = c
	}
}

// Timeout sets a deadline on unary RPC handlers (0 disables).
func Timeout(timeout time.Duration) ServerOption {
	return func(s *Server) {
		s.timeout = timeout
	}
}

// Middlewares appends service middleware. By default it applies to unary RPCs;
// enable EnableStreamMiddleware to also run the chain on streaming RPCs.
func Middlewares(middlewares ...middleware.Middleware) ServerOption {
	return func(s *Server) {
		s.middlewares = append(s.middlewares, middlewares...)
	}
}

// EnableStreamMiddleware runs the gkit middleware chain on streaming RPCs too.
// The chain executes once when the stream opens (with a nil request); the ctx
// it produces is propagated to the stream handler. Use only with middleware
// that is safe for streams (e.g. logging, recovery, metrics, auth). Off by
// default to preserve existing behavior.
func EnableStreamMiddleware() ServerOption {
	return func(s *Server) {
		s.streamMiddleware = true
	}
}

// SetMiddlewares is an alias for Middlewares.
func SetMiddlewares(middlewares ...middleware.Middleware) ServerOption {
	return Middlewares(middlewares...)
}

// UnaryInterceptor returns a ServerOption that sets the UnaryServerInterceptor for the server.
func UnaryInterceptor(in ...grpc.UnaryServerInterceptor) ServerOption {
	return func(s *Server) {
		s.unaryInterceptors = append(s.unaryInterceptors, in...)
	}
}

// StreamInterceptor returns a ServerOption that sets the StreamServerInterceptor for the server.
func StreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.streamInterceptors = append(s.streamInterceptors, in...)
	}
}

// Options appends extra grpc.ServerOption values.
func Options(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = append(s.grpcOpts, opts...)
	}
}

// Listener sets a pre-created listener (also used to derive endpoint).
func Listener(lis net.Listener) ServerOption {
	return func(s *Server) {
		s.listener = lis
	}
}

// DisableTransportCtx cancel default Transport context handling.
func DisableTransportCtx() ServerOption {
	return func(s *Server) {
		s.disableTransportCtx = true
	}
}

// DisableErrorInterceptor disables the default errorx status interceptor.
func DisableErrorInterceptor() ServerOption {
	return func(s *Server) {
		s.disableErrorInterceptor = true
	}
}
