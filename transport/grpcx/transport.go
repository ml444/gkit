package grpcx

import (
	"context"

	"google.golang.org/grpc"

	"github.com/ml444/gkit/pkg/header"
	"github.com/ml444/gkit/transport"
)

var _ transport.ITransport = (*Transport)(nil)

type Transport struct {
	transport.BaseTransport
}

func (tr *Transport) GetKind() transport.Kind {
	return transport.KindGRPC
}

func ClientTransport(ctx context.Context, method string, cc *grpc.ClientConn, header header.IHeader) transport.ITransport {
	return &Transport{
		BaseTransport: transport.BaseTransport{
			Endpoint:  cc.Target(),
			Operation: method,
			InHeader:  header,
			OutHeader: nil,
		},
	}
}
