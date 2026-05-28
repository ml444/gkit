package xds

import (
	"fmt"

	"google.golang.org/grpc"
	grpcxds "google.golang.org/grpc/xds"
)

// NewGRPCServer creates a gRPC server backed by the xDS control plane.
func NewGRPCServer(opts ...grpc.ServerOption) (*grpcxds.GRPCServer, error) {
	srv, err := grpcxds.NewGRPCServer(opts...)
	if err != nil {
		return nil, fmt.Errorf("xds: NewGRPCServer: %w", err)
	}
	return srv, nil
}
