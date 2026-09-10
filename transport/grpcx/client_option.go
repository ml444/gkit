package grpcx

import (
	"crypto/tls"
	"time"

	"github.com/ml444/gkit/discovery"
	"google.golang.org/grpc"
)

// ClientOption configures a gRPC client.
type ClientOption func(*Client)

// WithEndpoint sets the dial target (host:port, dns:///..., discovery:///service).
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.endpoint = endpoint
	}
}

// WithDiscovery enables discovery-based resolution for discovery:/// targets.
func WithDiscovery(dc *discovery.DiscoveryClient, serviceName string) ClientOption {
	return func(c *Client) {
		c.discovery = dc
		c.service = serviceName
	}
}

// WithTimeout sets both dial and unary call timeouts (0 disables each).
// Later WithDialTimeout or WithCallTimeout options override the respective value.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.timeout = d
		c.dialTimeout = d
	}
}

// WithTLSConfig sets transport TLS for direct and discovery targets.
func WithTLSConfig(cfg *tls.Config) ClientOption {
	return func(c *Client) {
		c.tlsConf = cfg
	}
}

// WithUnaryInterceptor appends unary client interceptors.
func WithUnaryInterceptor(in ...grpc.UnaryClientInterceptor) ClientOption {
	return func(c *Client) {
		c.unaryInterceptors = append(c.unaryInterceptors, in...)
	}
}

// WithDialOptions appends raw grpc.DialOption values.
func WithDialOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *Client) {
		c.dialOpts = append(c.dialOpts, opts...)
	}
}

// WithDialTimeout limits blocking connection setup; 0 disables the limit.
// Dialing remains non-blocking unless grpc.WithBlock is supplied.
func WithDialTimeout(d time.Duration) ClientOption { return func(c *Client) { c.dialTimeout = d } }

// WithCallTimeout sets the default unary RPC timeout; 0 disables it.
// A shorter caller deadline wins. Streams use only their caller's context.
func WithCallTimeout(d time.Duration) ClientOption { return func(c *Client) { c.timeout = d } }
