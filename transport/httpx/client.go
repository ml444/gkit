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

	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
)

// EncodeRequestFunc is request encode func.
type EncodeRequestFunc func(ctx context.Context, contentType string, in interface{}) (body []byte, err error)

// DecodeResponseFunc is response decode func.
type DecodeResponseFunc func(ctx context.Context, res *http.Response, out interface{}) error

// Client is an HTTP client.
type Client struct {
	target     *Target
	cc         *http.Client
	insecure   bool
	tlsConf    *tls.Config
	timeout    time.Duration
	endpoint   string
	userAgent  string
	encoder    EncodeRequestFunc
	decoder    DecodeResponseFunc
	transport  http.RoundTripper
	middleware []middleware.Middleware
}

// NewClient returns an HTTP client.
func NewClient(opts ...ClientOption) (*Client, error) {
	client := Client{
		timeout:   2000 * time.Millisecond,
		encoder:   DefaultRequestEncoder,
		decoder:   DefaultResponseDecoder,
		transport: http.DefaultTransport,
	}
	for _, o := range opts {
		o(&client)
	}
	client.cc = &http.Client{
		Timeout:   client.timeout,
		Transport: client.transport,
	}
	if client.tlsConf != nil {
		if tr, ok := client.transport.(*http.Transport); ok {
			tr.TLSClientConfig = client.tlsConf
		}
	}
	client.insecure = client.tlsConf == nil
	target, err := parseTarget(client.endpoint, client.insecure)
	if err != nil {
		return nil, err
	}
	client.target = target
	return &client, nil
}

// Invoke makes a rpc call procedure for remote service.
func (client *Client) Invoke(ctx context.Context, method, path string, args interface{}, reply interface{}, opts ...CallOption) error {
	var (
		contentType string
		body        io.Reader
	)
	c := defaultCallInfo(path)
	for _, o := range opts {
		o(&c)
	}
	contentType = c.reqHeader.Get("Content-Type")
	if args != nil {
		data, err := client.encoder(ctx, contentType, args)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	url0 := fmt.Sprintf("%s://%s%s", client.target.Scheme, client.target.Authority, path)
	req, err := http.NewRequest(method, url0, body)
	if err != nil {
		return err
	}
	if c.reqHeader != nil {
		req.Header = c.reqHeader
	}
	if client.userAgent != "" {
		req.Header.Set("User-Agent", client.userAgent)
	}
	ctx = transport.ToContext(ctx, &transport.Transport{
		Endpoint:  client.endpoint,
		InHeader:  (transport.Metadata)(req.Header),
		Operation: c.operation,
		Request:   req,
		//pathTemplate: c.pathTemplate,
	})
	return client.invoke(ctx, req, args, reply, c, opts...)
}

func (client *Client) invoke(ctx context.Context, req *http.Request, args interface{}, reply interface{}, c callInfo, opts ...CallOption) error {
	h := func(ctx context.Context, in interface{}) (interface{}, error) {
		res, err := client.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if err = client.decoder(ctx, res, reply); err != nil {
			return nil, err
		}
		return reply, nil
	}
	if len(client.middleware) > 0 {
		h = middleware.Chain(client.middleware...)(h)
	}
	_, err := h(ctx, args)
	return err
}

func (client *Client) Do(req *http.Request) (*http.Response, error) {
	if client.insecure {
		req.URL.Scheme = "http"
	} else {
		req.URL.Scheme = "https"
	}
	if client.endpoint != "" {
		req.URL.Host = client.endpoint
		req.Host = client.endpoint
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
