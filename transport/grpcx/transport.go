package grpcx

import (
	"context"

	"google.golang.org/grpc"

	"github.com/ml444/gkit/transport"
)

var _ transport.ITransport = (*Transport)(nil)

type Transport struct {
	endpoint  string
	operation string
	inMD      transport.MD
	outMD     transport.MD
}

func (tr *Transport) Kind() string {
	return "grpc"
}

func (tr *Transport) Endpoint() string {
	return tr.endpoint
}

func (tr *Transport) Path() string {
	return tr.operation
}

func (tr *Transport) In() transport.MD {
	return tr.inMD
}

func (tr *Transport) Out() transport.MD {
	return tr.outMD
}

// GetTransport returns grpc Transport from context.
func GetTransport(ctx context.Context) (*Transport, bool) {
	tr, ok := transport.FromContext(ctx)
	if !ok {
		return nil, false
	}
	gtr, ok := tr.(*Transport)
	return gtr, ok
}

func ClientTransport(ctx context.Context, method string, cc *grpc.ClientConn, header transport.MD) transport.ITransport {
	return &Transport{
		endpoint:  cc.Target(),
		operation: method,
		inMD:      header,
		outMD:     nil,
	}
}
