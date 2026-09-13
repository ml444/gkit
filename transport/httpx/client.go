package httpx

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/ml444/gkit/discovery"
	"github.com/ml444/gkit/middleware"
	"github.com/ml444/gkit/transport/httpx/coder/form"
)

// EncodeRequestFunc is request encode func.
type EncodeRequestFunc func(ctx context.Context, contentType string, in interface{}) (body []byte, err error)

// DecodeResponseFunc is response decode func.
type DecodeResponseFunc func(ctx context.Context, res *http.Response, out interface{}) error

// Client is an HTTP client.
type Client struct {
	configErr     error
	responseHooks []ResponseHeadersHook
	target        *Target
	cc            *http.Client
	insecure      bool
	tlsConf       *tls.Config
	timeout       time.Duration
	endpoint      string
	discovery     *discovery.DiscoveryClient
	service       string
	userAgent     string
	encoder       EncodeRequestFunc
	decoder       DecodeResponseFunc
	transport     http.RoundTripper
	middleware    []middleware.Middleware
}

// NewClient returns an HTTP client.
func NewClient(opts ...ClientOption) (*Client, error) {
	client := Client{
		timeout: 2000 * time.Millisecond,
		encoder: DefaultRequestEncoder,
		decoder: DefaultResponseDecoder,
	}
	for _, o := range opts {
		if o == nil {
			return nil, fmt.Errorf("httpx: nil client option")
		}
		o(&client)
	}
	if client.configErr != nil {
		return nil, client.configErr
	}
	if client.timeout < 0 {
		return nil, fmt.Errorf("httpx: timeout must not be negative")
	}
	if client.encoder == nil || client.decoder == nil {
		return nil, fmt.Errorf("httpx: encoder and decoder are required")
	}
	// Resolve the RoundTripper without mutating the process-wide
	// http.DefaultTransport. When a TLS config is supplied we clone a transport
	// so other clients sharing DefaultTransport are not affected.
	if client.transport == nil {
		if client.tlsConf != nil {
			base, ok := http.DefaultTransport.(*http.Transport)
			if !ok {
				return nil, fmt.Errorf("httpx: TLS config requires an http.Transport; provide WithTransport")
			}
			tr := base.Clone()
			tr.TLSClientConfig = client.tlsConf
			client.transport = tr
		} else {
			client.transport = http.DefaultTransport
		}
	} else if client.tlsConf != nil {
		if tr, ok := client.transport.(*http.Transport); ok {
			cloned := tr.Clone()
			cloned.TLSClientConfig = client.tlsConf
			client.transport = cloned
		}
	}
	client.cc = &http.Client{
		Timeout:   client.timeout,
		Transport: client.transport,
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
	if client.discovery != nil {
		if client.service == "" {
			return nil, fmt.Errorf("httpx: discovery service name is required")
		}
		if client.target.Authority == "" {
			client.target.Authority = "discovery"
		}
	} else if client.target.DiscoveryService != "" {
		return nil, fmt.Errorf("httpx: discovery target requires WithDiscovery")
	}
	return &client, nil
}

// Invoke encodes args and decodes the response within the configured middleware chain.
func (client *Client) Invoke(ctx context.Context, method, path string, args interface{}, reply interface{}, opts ...CallOption) error {
	if err := client.validateInvocation(path); err != nil {
		return err
	}
	c, err := parseCallOptions(path, opts)
	if err != nil {
		return err
	}
	var body io.Reader
	if args != nil {
		data, err := client.encoder(ctx, c.reqHeader.Get("Content-Type"), args)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	ctx, req, err := client.prepareInvocation(ctx, method, path, body, c)
	if err != nil {
		return err
	}
	_, err = client.execute(ctx, req, args, c, responseHandler{
		consume: func(ctx context.Context, res *http.Response) (interface{}, error) {
			if err := client.decoder(ctx, res, reply); err != nil {
				return nil, err
			}
			return reply, nil
		},
	}, false)
	return err
}

// Do sends a copy of req. A configured endpoint overrides its destination;
// without an endpoint, the request's own URL (including HTTPS) is preserved.
func (client *Client) Do(req *http.Request) (*http.Response, error) {
	if req == nil || req.URL == nil {
		return nil, fmt.Errorf("httpx: request and URL are required")
	}
	req = req.Clone(req.Context())
	if target := client.target; target != nil && target.Authority != "" {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Authority
		req.Host = target.Authority
		if target.Endpoint != "" {
			escaped := JoinPath("/"+target.Endpoint, req.URL.EscapedPath())
			decoded, err := url.PathUnescape(escaped)
			if err != nil {
				return nil, err
			}
			req.URL.Path, req.URL.RawPath = decoded, escaped
		}
	}
	if client.discovery != nil {
		inst, err := client.discovery.GetServiceInstance(req.Context(), client.service)
		if err != nil {
			return nil, err
		}
		if hd, ok := req.Context().Value(instanceHolderKey{}).(*instanceHolder); ok {
			hd.inst = inst
		}
		req.URL.Host = net.JoinHostPort(inst.GetAddress(), strconv.Itoa(inst.GetPort()))
		req.Host = req.URL.Host
	}
	return client.cc.Do(req)
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

type instanceHolderKey struct{}

type instanceHolder struct {
	inst discovery.ServiceInstancer
}

func (client *Client) updateDiscoveryStatus(ctx context.Context, holder *instanceHolder, success bool) {
	if client == nil || client.discovery == nil || holder == nil {
		return
	}
	if holder.inst == nil {
		return
	}
	client.discovery.UpdateLoadBalancerStatus(ctx, holder.inst, success)
}

// Target is resolver target
type Target struct {
	Scheme           string
	Authority        string
	Endpoint         string
	DiscoveryService string
}

func parseTarget(endpoint string, insecure bool) (*Target, error) {
	scheme := "http"
	if !insecure {
		scheme = "https"
	}
	if endpoint == "" {
		return &Target{Scheme: scheme}, nil
	}
	if !strings.Contains(endpoint, "://") {
		endpoint = scheme + "://" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	if u.User != nil || u.RawQuery != "" || u.ForceQuery || strings.Contains(endpoint, "#") {
		return nil, fmt.Errorf("httpx: endpoint must not contain credentials, query or fragment")
	}
	if u.Scheme == "discovery" {
		service := strings.TrimPrefix(u.Path, "/")
		if service == "" || u.Host != "" {
			return nil, fmt.Errorf("httpx: expected discovery:///service")
		}
		return &Target{Scheme: scheme, Authority: "discovery", DiscoveryService: service}, nil
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("httpx: unsupported endpoint scheme %q", u.Scheme)
	}
	if u.Hostname() == "" {
		return nil, fmt.Errorf("httpx: endpoint host is required")
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("httpx: invalid endpoint port %q", port)
		}
	}
	return &Target{Scheme: u.Scheme, Authority: u.Host, Endpoint: strings.TrimPrefix(u.EscapedPath(), "/")}, nil
}

var reg = regexp.MustCompile(`{[.\w]+}`)

// EncodeURL encodes path parameters and query values.
// Deprecated: use EncodeURLWithError to detect missing parameters and encoding errors.
func EncodeURL(pathTemplate string, msg interface{}, needQuery bool) string {
	path, _ := encodeURL(pathTemplate, msg, needQuery, false)
	return path
}

// EncodeURLWithError escapes each ordinary placeholder as one path segment and
// reports missing parameters or encoding errors. Complex path templates are not supported.
func EncodeURLWithError(pathTemplate string, msg interface{}, needQuery bool) (string, error) {
	return encodeURL(pathTemplate, msg, needQuery, true)
}

func encodeURL(pathTemplate string, msg interface{}, needQuery, strict bool) (string, error) {
	isNil := msg == nil || (reflect.ValueOf(msg).Kind() == reflect.Ptr && reflect.ValueOf(msg).IsNil())
	if isNil && !strict {
		return pathTemplate, nil
	}
	queryParams, err := form.EncodeValues(msg)
	if err != nil && strict {
		return "", err
	}
	pathParams := make(map[string]struct{})
	var missing string
	path := reg.ReplaceAllStringFunc(pathTemplate, func(in string) string {
		key := in[1 : len(in)-1]
		// Protobuf templates can use proto field names as well as JSON names.
		if m, ok := msg.(proto.Message); ok && !isNil {
			md := m.ProtoReflect().Descriptor()
			parts := strings.Split(key, ".")
			for i, part := range parts {
				fd := md.Fields().ByJSONName(part)
				if fd == nil {
					fd = md.Fields().ByTextName(part)
				}
				if fd == nil {
					break
				}
				parts[i] = fd.JSONName()
				if i < len(parts)-1 {
					if fd.Message() == nil {
						break
					}
					md = fd.Message()
				}
			}
			key = strings.Join(parts, ".")
		}
		pathParams[key] = struct{}{}
		values := queryParams[key]
		if len(values) != 1 || values[0] == "" {
			missing = key
		}
		return url.PathEscape(queryParams.Get(key))
	})
	if strict && missing != "" {
		return "", fmt.Errorf("httpx: missing or non-scalar path parameter %q", missing)
	}
	if strict && strings.ContainsAny(path, "{}") {
		return "", fmt.Errorf("httpx: unsupported path template %q", pathTemplate)
	}
	u, err := url.Parse(path)
	if err != nil {
		if strict {
			return "", err
		}
		return path, nil
	}
	query, queryErr := url.ParseQuery(u.RawQuery)
	if queryErr != nil && strict {
		return "", queryErr
	}
	if needQuery {
		for key, values := range queryParams {
			if _, used := pathParams[key]; used {
				continue
			}
			for _, value := range values {
				query.Add(key, value)
			}
		}
	} else if m, ok := msg.(proto.Message); ok && !isNil {
		mask, err := url.ParseQuery(form.EncodeFieldMask(m.ProtoReflect()))
		if err != nil && strict {
			return "", err
		}
		for key, values := range mask {
			query[key] = values
		}
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}
