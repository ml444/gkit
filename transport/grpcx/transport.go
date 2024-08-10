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

func (tr *Transport) Method() string {
	return tr.operation
}

func (tr *Transport) InHeader() transport.MD {
	return tr.inMD
}

func (tr *Transport) OutHeader() transport.MD {
	return tr.outMD
}

func ClientTransport(ctx context.Context, method string, cc *grpc.ClientConn, header transport.MD) transport.ITransport {
	return &Transport{
		endpoint:  cc.Target(),
		operation: method,
		inMD:      header,
		outMD:     nil,
	}
}
