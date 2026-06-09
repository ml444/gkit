package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport"
	"github.com/ml444/gkit/transport/httpx/coder/form"
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
	discovery  *discovery.DiscoveryClient
	service    string
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
	if client.service == "" && client.target != nil && client.target.DiscoveryService != "" {
		client.service = client.target.DiscoveryService
	}
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
	ctx = transport.ToContext(ctx, &Transport{
		endpoint:     client.endpoint,
		path:         c.operation,
		inMD:         transport.New(req.Header),
		pathTemplate: c.pathTemplate,
	})
	return client.invoke(ctx, req, args, reply, c, opts...)
}

func (client *Client) invoke(ctx context.Context, req *http.Request, args interface{}, reply interface{}, c callInfo, opts ...CallOption) error {
	h := func(ctx context.Context, in interface{}) (interface{}, error) {
		res, req, err := client.Do(req.WithContext(ctx))
		if err != nil {
			client.updateDiscoveryStatus(req.Context(), false)
			return nil, err
		}
		defer res.Body.Close()
		if err = client.decoder(ctx, res, reply); err != nil {
			client.updateDiscoveryStatus(req.Context(), false)
			return nil, err
		}
		if c.onResponse != nil {
			if err = c.onResponse(res); err != nil {
				client.updateDiscoveryStatus(req.Context(), false)
				return nil, err
			}
		}
		client.updateDiscoveryStatus(req.Context(), res.StatusCode < 500)
		return reply, nil
	}
	if len(client.middleware) > 0 {
		h = middleware.Chain(client.middleware...)(h)
	}
	_, err := h(ctx, args)
	return err
}

func (client *Client) Do(req *http.Request) (*http.Response, *http.Request, error) {
	if client.insecure {
		req.URL.Scheme = "http"
	} else {
		req.URL.Scheme = "https"
	}
	if client.discovery != nil {
		svc := client.service
		if svc == "" && client.target != nil {
			svc = client.target.DiscoveryService
		}
		if svc == "" {
			return nil, req, fmt.Errorf("httpx: discovery enabled but service name is empty")
		}
		inst, err := client.discovery.GetServiceInstance(req.Context(), svc)
		if err != nil {
			return nil, req, err
		}
		req = req.WithContext(context.WithValue(req.Context(), discoveryInstanceKey{}, inst))
		req.URL.Host = fmt.Sprintf("%s:%d", inst.GetAddress(), inst.GetPort())
		req.Host = req.URL.Host
	} else if client.endpoint != "" {
		req.URL.Host = client.endpoint
		req.Host = client.endpoint
	}
	res, err := client.cc.Do(req)
	return res, req, err
}

// Close tears down the Transport and all underlying connections.
func (client *Client) Close() error {
	if client == nil || client.transport == nil {
		return nil
	}
	if tr, ok := client.transport.(interface{ CloseIdleConnections() }); ok {
		tr.CloseIdleConnections()
	}
	return nil
}

type discoveryInstanceKey struct{}

func (client *Client) updateDiscoveryStatus(ctx context.Context, success bool) {
	if client == nil || client.discovery == nil {
		return
	}
	v := ctx.Value(discoveryInstanceKey{})
	if v == nil {
		return
	}
	inst, ok := v.(discovery.ServiceInstancer)
	if !ok || inst == nil {
		return
	}
	client.discovery.UpdateLoadBalancerStatus(ctx, inst, success)
}

// Target is resolver target
type Target struct {
	Scheme           string
	Authority        string
	Endpoint         string
	DiscoveryService string
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
	if u.Scheme == "discovery" {
		scheme := "https"
		if insecure {
			scheme = "http"
		}
		service := ""
		if len(u.Path) > 1 {
			service = u.Path[1:]
		}
		return &Target{
			Scheme:           scheme,
			Authority:        "discovery",
			DiscoveryService: service,
		}, nil
	}
	target := &Target{Scheme: u.Scheme, Authority: u.Host}
	if len(u.Path) > 1 {
		target.Endpoint = u.Path[1:]
	}
	return target, nil
}

var reg = regexp.MustCompile(`{[\\.\w]+}`)

// EncodeURL encode proto message to url path.
func EncodeURL(pathTemplate string, msg interface{}, needQuery bool) string {
	if msg == nil || (reflect.ValueOf(msg).Kind() == reflect.Ptr && reflect.ValueOf(msg).IsNil()) {
		return pathTemplate
	}
	queryParams, _ := form.EncodeValues(msg)
	pathParams := make(map[string]struct{})
	path := reg.ReplaceAllStringFunc(pathTemplate, func(in string) string {
		// it's unreachable because the reg means that must have more than one char in {}
		// if len(in) < 4 { //nolint:gomnd // **  explain the 4 number here :-) **
		//	return in
		// }
		key := in[1 : len(in)-1]
		pathParams[key] = struct{}{}
		return queryParams.Get(key)
	})
	if !needQuery {
		if v, ok := msg.(proto.Message); ok {
			if query := form.EncodeFieldMask(v.ProtoReflect()); query != "" {
				return path + "?" + query
			}
		}
		return path
	}
	if len(queryParams) > 0 {
		for key := range pathParams {
			delete(queryParams, key)
		}
		if query := queryParams.Encode(); query != "" {
			path += "?" + query
		}
	}
	return path
}
