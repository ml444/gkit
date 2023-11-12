package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ml444/gkit/transport"
)

// EncodeRequestFunc is request encode func.
type EncodeRequestFunc func(ctx context.Context, contentType string, in interface{}) (body []byte, err error)

// DecodeResponseFunc is response decode func.
type DecodeResponseFunc func(ctx context.Context, res *http.Response, out interface{}) error

// ClientOption is HTTP client option.
type ClientOption func(*clientOptions)

// Client is an HTTP transport client.
type clientOptions struct {
	ctx       context.Context
	tlsConf   *tls.Config
	timeout   time.Duration
	endpoint  string
	userAgent string
	encoder   EncodeRequestFunc
	decoder   DecodeResponseFunc
	transport http.RoundTripper
	//middleware   []middleware.Middleware
	block      bool
	subsetSize int
}

// WithSubset with client disocvery subset size.
// zero value means subset globalMiddleware disabled
func WithSubset(size int) ClientOption {
	return func(o *clientOptions) {
		o.subsetSize = size
	}
}

// WithTransport with client transport.
func WithTransport(trans http.RoundTripper) ClientOption {
	return func(o *clientOptions) {
		o.transport = trans
	}
}

// WithTimeout with client request timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = d
	}
}

// WithUserAgent with client user agent.
func WithUserAgent(ua string) ClientOption {
	return func(o *clientOptions) {
		o.userAgent = ua
	}
}

// WithMiddleware with client middleware.
//func WithMiddleware(m ...middleware.Middleware) ClientOption {
//	return func(o *clientOptions) {
//		o.middleware = m
//	}
//}

// WithEndpoint with client addr.
func WithEndpoint(endpoint string) ClientOption {
	return func(o *clientOptions) {
		o.endpoint = endpoint
	}
}

// WithRequestEncoder with client request encoder.
func WithRequestEncoder(encoder EncodeRequestFunc) ClientOption {
	return func(o *clientOptions) {
		o.encoder = encoder
	}
}

// WithResponseDecoder with client response decoder.
func WithResponseDecoder(decoder DecodeResponseFunc) ClientOption {
	return func(o *clientOptions) {
		o.decoder = decoder
	}
}

// WithBlock with client block.
func WithBlock() ClientOption {
	return func(o *clientOptions) {
		o.block = true
	}
}

// WithTLSConfig with tls config.
func WithTLSConfig(c *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.tlsConf = c
	}
}

// Client is an HTTP client.
type Client struct {
	opts     clientOptions
	target   *Target
	cc       *http.Client
	insecure bool
}

// NewClient returns an HTTP client.
func NewClient(ctx context.Context, opts ...ClientOption) (*Client, error) {
	options := clientOptions{
		ctx:        ctx,
		timeout:    2000 * time.Millisecond,
		encoder:    DefaultRequestEncoder,
		decoder:    DefaultResponseDecoder,
		transport:  http.DefaultTransport,
		subsetSize: 25,
	}
	for _, o := range opts {
		o(&options)
	}
	if options.tlsConf != nil {
		if tr, ok := options.transport.(*http.Transport); ok {
			tr.TLSClientConfig = options.tlsConf
		}
	}
	insecure := options.tlsConf == nil
	target, err := parseTarget(options.endpoint, insecure)
	if err != nil {
		return nil, err
	}
	return &Client{
		opts:     options,
		target:   target,
		insecure: insecure,
		cc: &http.Client{
			Timeout:   options.timeout,
			Transport: options.transport,
		},
	}, nil
}

// Invoke makes a rpc call procedure for remote service.
func (client *Client) Invoke(ctx context.Context, method, path string, args interface{}, reply interface{}, opts ...CallOption) error {
	var (
		contentType string
		body        io.Reader
	)
	c := defaultCallInfo(path)
	for _, o := range opts {
		if err := o.before(&c); err != nil {
			return err
		}
	}
	if args != nil {
		data, err := client.opts.encoder(ctx, c.contentType, args)
		if err != nil {
			return err
		}
		contentType = c.contentType
		body = bytes.NewReader(data)
	}
	url := fmt.Sprintf("%s://%s%s", client.target.Scheme, client.target.Authority, path)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", c.contentType)
	}
	if client.opts.userAgent != "" {
		req.Header.Set("User-Agent", client.opts.userAgent)
	}
	ctx = transport.ToContext(ctx, &transport.Transport{
		Endpoint:  client.opts.endpoint,
		InHeader:  (transport.Metadata)(req.Header),
		Operation: c.operation,
		Request:   req,
		//pathTemplate: c.pathTemplate,
	})
	return client.invoke(ctx, req, args, reply, c, opts...)
}

func (client *Client) invoke(ctx context.Context, req *http.Request, args interface{}, reply interface{}, c callInfo, opts ...CallOption) error {
	h := func(ctx context.Context, in interface{}) (interface{}, error) {
		res, err := client.do(req.WithContext(ctx))
		if res != nil {
			cs := csAttempt{res: res}
			for _, o := range opts {
				o.after(&c, &cs)
			}
		}
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if err = client.opts.decoder(ctx, res, reply); err != nil {
			return nil, err
		}
		return reply, nil
	}

	_, err := h(ctx, args)
	return err
}

// Do send an HTTP request and decodes the body of response into target.
// returns an error (of type *Error) if the response status code is not 2xx.
func (client *Client) Do(req *http.Request, opts ...CallOption) (*http.Response, error) {
	c := defaultCallInfo(req.URL.Path)
	for _, o := range opts {
		if err := o.before(&c); err != nil {
			return nil, err
		}
	}

	return client.do(req)
}

func (client *Client) do(req *http.Request) (*http.Response, error) {
	if client.insecure {
		req.URL.Scheme = "http"
	} else {
		req.URL.Scheme = "https"
	}
	if client.opts.endpoint != "" {
		req.URL.Host = client.opts.endpoint
		req.Host = client.opts.endpoint
		return client.cc.Do(req)
	}

	return client.cc.Do(req)
}

// Close tears down the Transport and all underlying connections.
func (client *Client) Close() error {
	return nil
}

// Target is resolver target
type Target struct {
	Scheme    string
	Authority string
	Endpoint  string
}

func parseTarget(endpoint string, insecure bool) (*Target, error) {
	if !strings.Contains(endpoint, "://") {
		if insecure {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	target := &Target{Scheme: u.Scheme, Authority: u.Host}
	if len(u.Path) > 1 {
		target.Endpoint = u.Path[1:]
	}
	return target, nil
}
