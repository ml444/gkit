package xds

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	xdscreds "google.golang.org/grpc/credentials/xds"

	"github.com/ml444/gkit/middleware/response"
)

// ClientOption configures an xDS gRPC client.
type ClientOption func(*clientConfig)

type clientConfig struct {
	dialOpts []grpc.DialOption
}

// WithDialOptions appends grpc.DialOption values.
func WithDialOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *clientConfig) {
		c.dialOpts = append(c.dialOpts, opts...)
	}
}

// NewClient dials an xDS target (e.g. xds:///listener-name).
// Requires xDS bootstrap via GRPC_XDS_BOOTSTRAP environment variable.
func NewClient(target string, opts ...ClientOption) (*grpc.ClientConn, error) {
	cfg := &clientConfig{}
	for _, o := range opts {
		o(cfg)
	}
	creds, err := xdscreds.NewClientCredentials(
		xdscreds.ClientOptions{FallbackCreds: insecure.NewCredentials()},
	)
	if err != nil {
		return nil, err
	}
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithChainUnaryInterceptor(response.ClientErrorInterceptor),
	}
	dialOpts = append(dialOpts, cfg.dialOpts...)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return grpc.DialContext(ctx, target, dialOpts...)
}
