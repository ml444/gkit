package grpcx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
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
	"github.com/ml444/gkit/log"
	"github.com/ml444/gkit/middleware/response"
	"github.com/ml444/gkit/transport/grpcx/resolver"
	"github.com/ml444/gkit/transport/grpcx/xds"
)

// Client is a gRPC client with optional discovery integration.
type Client struct {
	conn              *grpc.ClientConn
	endpoint          string
	service           string
	discovery         *discovery.DiscoveryClient
	timeout           time.Duration // unary call timeout
	dialTimeout       time.Duration
	resolver          *resolver.Builder
	tlsConf           *tls.Config
	unaryInterceptors []grpc.UnaryClientInterceptor
	dialOpts          []grpc.DialOption
}

// NewClient creates a gRPC client connection.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{
		timeout:     10 * time.Second,
		dialTimeout: 10 * time.Second,
	}
	for _, o := range opts {
		o(c)
	}
	if c.timeout < 0 || c.dialTimeout < 0 {
		return nil, fmt.Errorf("grpcx: timeouts must not be negative")
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
				c.timeoutInterceptor(),
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
		c.resolver = resolver.NewBuilder(c.discovery)
		dialOpts = append(dialOpts, grpc.WithResolvers(c.resolver))
		const serviceConfig = `{"loadBalancingConfig":[{"round_robin":{}}]}`
		dialOpts = append(dialOpts, grpc.WithDefaultServiceConfig(serviceConfig))
	}
	dialOpts = append(dialOpts, c.dialOpts...)

	ctx := context.Background()
	if c.dialTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.dialTimeout)
		defer cancel()
	}
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
		if u.Host != "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(endpoint, "#") {
			return "", "", fmt.Errorf("grpcx: expected discovery:///service without credentials, query or fragment")
		}
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
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) timeoutInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if c.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, c.timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (c *Client) discoveryFeedbackInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if c.discovery == nil || c.service == "" {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		var p peer.Peer
		callOpts := append([]grpc.CallOption{grpc.Peer(&p)}, opts...)
		err := invoker(ctx, method, req, reply, cc, callOpts...)
		if p.Addr == nil || c.resolver == nil {
			log.Debugf("grpcx: skipping instance feedback for %s: peer or resolver unavailable", method)
			return err
		}
		inst, ok := c.resolver.GetInstance(c.service, p.Addr.String())
		if !ok {
			log.Debugf("grpcx: skipping instance feedback for %s: unknown peer %s", method, p.Addr)
			return err
		}
		success := err == nil || (status.Code(err) != codes.Unavailable && status.Code(err) != codes.DeadlineExceeded)
		c.discovery.UpdateLoadBalancerStatus(ctx, inst, success)
		return err
	}
}

func instanceFromPeer(ctx context.Context) (discovery.ServiceInstancer, error) {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr == nil {
		return nil, errors.New("grpc peer is nil")
	}
	inst, ok := resolver.GetInstanceByAddr(p.Addr.String())
	if !ok {
		return nil, fmt.Errorf("instance not found for peer addr: %s", p.Addr.String())
	}
	return inst, nil
}

// NewXDSConn dials an xDS target. Deprecated: use xds.NewClient directly.
func NewXDSConn(dsn string) (*grpc.ClientConn, error) {
	return xds.NewClient(dsn)
}
