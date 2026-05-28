package grpcx

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/ml444/gkit/discovery"
)

// RegisterOption configures service registration.
type RegisterOption func(*registerConfig)

type registerConfig struct {
	id       string
	name     string
	version  string
	metadata map[string]string
}

// RegisterID sets the service instance ID.
func RegisterID(id string) RegisterOption {
	return func(c *registerConfig) {
		c.id = id
	}
}

// RegisterVersion sets the service version.
func RegisterVersion(version string) RegisterOption {
	return func(c *registerConfig) {
		c.version = version
	}
}

// RegisterMetadata sets instance metadata.
func RegisterMetadata(md map[string]string) RegisterOption {
	return func(c *registerConfig) {
		c.metadata = md
	}
}

func (s *Server) buildServiceInstance(opts ...RegisterOption) (*discovery.ServiceInstance, error) {
	endpoint, err := s.Endpoint()
	if err != nil {
		return nil, fmt.Errorf("grpcx: endpoint: %w", err)
	}
	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, fmt.Errorf("grpcx: split endpoint %q: %w", endpoint, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("grpcx: parse port %q: %w", portStr, err)
	}
	cfg := registerConfig{
		name: s.name,
		id:   fmt.Sprintf("%s-%s-%d", s.name, host, port),
	}
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.name == "" {
		cfg.name = s.name
	}
	return &discovery.ServiceInstance{
		ID:       cfg.id,
		Name:     cfg.name,
		Version:  cfg.version,
		Address:  host,
		Port:     port,
		Metadata: cfg.metadata,
	}, nil
}

// RegisterDiscovery registers this server with a service registry.
func (s *Server) RegisterDiscovery(ctx context.Context, reg discovery.ServiceRegistry, opts ...RegisterOption) error {
	inst, err := s.buildServiceInstance(opts...)
	if err != nil {
		return err
	}
	return reg.Register(ctx, inst)
}

// DeregisterDiscovery removes this server from a service registry.
func (s *Server) DeregisterDiscovery(ctx context.Context, reg discovery.ServiceRegistry, opts ...RegisterOption) error {
	inst, err := s.buildServiceInstance(opts...)
	if err != nil {
		return err
	}
	return reg.Deregister(ctx, inst)
}
