package httpx

import (
	"context"
	"net/http"

	"github.com/ml444/gkit/transport"
)

var _ transport.ITransport = (*Transport)(nil)

type Transport struct {
	endpoint string
	path     string
	// pathTemplate is the mux template form, e.g. /path/{id}
	pathTemplate string
	inMD     transport.MD
	outMD    transport.MD

	req *http.Request
}

func (tr *Transport) Kind() string {
	return "http"
}

func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

func (tr *Transport) Path() string {
	return tr.path
}

func (tr *Transport) PathTemplate() string {
	return tr.pathTemplate
}

func (tr *Transport) In() transport.MD {
	return tr.inMD
}

func (tr *Transport) Out() transport.MD {
	return tr.outMD
}

func (tr *Transport) Request() *http.Request {
	return tr.req
}

func (tr *Transport) SetEndpoint(endpoint string) {
	tr.endpoint = endpoint
}

func (tr *Transport) SetPath(path string) {
	tr.path = path
}

func (tr *Transport) SetRequestHeader(headers http.Header) {
	tr.inMD = transport.MD(headers)
}

func (tr *Transport) SetResponseHeader(headers http.Header) {
	tr.outMD = transport.MD(headers)
}

// SetOperation sets the transport Operation.
func SetOperation(ctx context.Context, op string) {
	if tr, ok := transport.FromContext(ctx); ok {
		if httpTr, ok := tr.(*Transport); ok {
			httpTr.path = op
		}
	}
}

func ClientTransport(r *http.Request) transport.ITransport {
	return &Transport{
		endpoint: r.URL.String(),
		path:     r.URL.Path,
		// best-effort: at client side we don't have a mux template
		pathTemplate: r.URL.Path,
		inMD:     transport.MD(r.Header),
		req:      r,
	}
}

func GetTransport(ctx context.Context) (*Transport, bool) {
	tr, ok := transport.FromContext(ctx)
	if !ok {
		return nil, false
	}
	httpTr, okk := tr.(*Transport)
	return httpTr, okk
}
