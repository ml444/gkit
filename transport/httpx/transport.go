package httpx

import (
	"context"
	"net/http"

	"github.com/ml444/gkit/pkg/header"
	"github.com/ml444/gkit/transport"
)

var _ transport.ITransport = (*Transport)(nil)

type Transport struct {
	transport.BaseTransport
}

func (tr *Transport) GetKind() transport.Kind {
	return transport.KindHTTP
}

// SetOperation sets the transport Operation.
func SetOperation(ctx context.Context, op string) {
	if tr, ok := transport.FromContext(ctx); ok {
		if httpTr, ok := tr.(*Transport); ok {
			httpTr.Operation = op
		}
	}
}

func ClientTransport(r *http.Request) transport.ITransport {
	return &Transport{
		BaseTransport: transport.BaseTransport{
			Endpoint:  r.URL.String(),
			Operation: r.URL.Path,
			InHeader:  header.Header(r.Header),
			OutHeader: nil,
		},
	}
}
