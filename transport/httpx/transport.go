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

func (tr *Transport) Method() string {
	return tr.path
}

func (tr *Transport) InHeader() transport.MD {
	return tr.inMD
}

func (tr *Transport) OutHeader() transport.MD {
	return tr.outMD
}
func (tr *Transport) Request() *http.Request {
	return tr.req
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
		inMD:     transport.MD(r.Header),
		req:      r,
	}
}
