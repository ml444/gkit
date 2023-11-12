package grpcx

import (
	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServerOption is gRPC server option.
type ServerOption func(o *Server)

func Debug() ServerOption {
	return func(s *Server) {
		s.debug = true
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

// EnableHealth Checks server.
func EnableHealth() ServerOption {
	return func(s *Server) {
		s.enableHealth = true
	}
}

// Credentials with server credentials.
func Credentials(creds credentials.TransportCredentials) ServerOption {
	return func(s *Server) {
		s.creds = creds
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
		s.grpcOpts = append(s.grpcOpts, grpc.ChainUnaryInterceptor(in...))
	}
}

// StreamInterceptor returns a ServerOption that sets the StreamServerInterceptor for the server.
func StreamInterceptor(in ...grpc.StreamServerInterceptor) ServerOption {
	return func(s *Server) {
		s.grpcOpts = append(s.grpcOpts, grpc.ChainStreamInterceptor(in...))
	}
}

// Options with grpc options.
func Options(opts ...grpc.ServerOption) ServerOption {
	return func(s *Server) {
		s.grpcOpts = opts
	}
}
