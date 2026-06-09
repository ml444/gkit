package grpcx

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/middleware/response"
	discoveryresolver "github.com/ml444/gkit/transport/grpcx/resolver"
	"github.com/ml444/gkit/transport/grpcx/xds"
)

// Client is a gRPC client with optional discovery integration.
type Client struct {
	conn              *grpc.ClientConn
	endpoint          string
	service           string
	discovery         *discovery.DiscoveryClient
	timeout           time.Duration
	tlsConf           *tls.Config
	unaryInterceptors []grpc.UnaryClientInterceptor
	dialOpts          []grpc.DialOption
}

// NewClient creates a gRPC client connection.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		timeout: 10 * time.Second,
	}
	for _, o := range opts {
		o(c)
	}
	target, service, err := parseClientTarget(c.endpoint, c.service)
	if err != nil {
		return nil, err
	}
	c.service = service

	dialOpts := []grpc.DialOption{
		grpc.WithChainUnaryInterceptor(append(
			[]grpc.UnaryClientInterceptor{
				response.ClientErrorInterceptor,
				c.discoveryFeedbackInterceptor(),
			},
			c.unaryInterceptors...,
		)...),
	}
	if c.tlsConf != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConf)))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	if strings.HasPrefix(target, "discovery:///") {
		if c.discovery == nil {
			return nil, fmt.Errorf("grpcx: discovery target requires WithDiscovery")
		}
		discoveryresolver.Register(c.discovery)
		const serviceConfig = `{"loadBalancingConfig":[{"round_robin":{}}]}`
		dialOpts = append(dialOpts, grpc.WithDefaultServiceConfig(serviceConfig))
	}
	dialOpts = append(dialOpts, c.dialOpts...)

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, target, dialOpts...)
	if err != nil {
		return nil, err
	}
	c.conn = conn
	return c, nil
}

func parseClientTarget(endpoint, service string) (target, serviceName string, err error) {
	if endpoint == "" {
		return "", "", fmt.Errorf("grpcx: endpoint is required")
	}
	if !strings.Contains(endpoint, "://") {
		endpoint = "passthrough:///" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}
	if u.Scheme == "discovery" {
		svc := strings.TrimPrefix(u.Path, "/")
		if service != "" {
			svc = service
		}
		if svc == "" {
			return "", "", fmt.Errorf("grpcx: discovery service name is empty")
		}
		return fmt.Sprintf("discovery:///%s", svc), svc, nil
	}
	if service != "" {
		return endpoint, service, nil
	}
	return endpoint, "", nil
}

// Conn returns the underlying ClientConn.
func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) discoveryFeedbackInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := invoker(ctx, method, req, reply, cc, opts...)
		if c.discovery == nil || c.service == "" {
			return err
		}
		inst := instanceFromPeer(ctx, c.discovery, c.service)
		if inst != nil {
			success := err == nil || (status.Code(err) != codes.Unavailable && status.Code(err) != codes.DeadlineExceeded)
			c.discovery.UpdateLoadBalancerStatus(ctx, inst, success)
		}
		return err
	}
}

func instanceFromPeer(ctx context.Context, dc *discovery.DiscoveryClient, service string) discovery.ServiceInstancer {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return nil
	}
	peerHost, peerPort, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return nil
	}
	instances, err := dc.GetAllInstances(ctx, service)
	if err != nil {
		return nil
	}
	for _, inst := range instances {
		if inst.GetAddress() == peerHost && fmt.Sprintf("%d", inst.GetPort()) == peerPort {
			return inst
		}
	}
	return nil
}

// NewXDSConn dials an xDS target. Deprecated: use xds.NewClient directly.
func NewXDSConn(dsn string) (*grpc.ClientConn, error) {
	return xds.NewClient(dsn)
}
