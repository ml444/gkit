package grpcx

import (
	"crypto/tls"

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

func EnableXDS() ServerOption {
	return func(s *Server) {
		s.enableXDS = true
	}
}

// EnableHealth Checks server.
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

// Options with grpc options.
func Options(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}

// DisableTransportCtx cancel default Transport context handling
func DisableTransportCtx() ServerOption {
	return func(s *Server) {
		s.disableTransportCtx = true
	}
}
