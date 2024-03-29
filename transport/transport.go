package transport

import (
	"context"
	"net/http"

	"google.golang.org/grpc"

	"github.com/ml444/gkit/pkg/header"
)

// Type defines the type of Transport
type Type string

func (k Type) String() string { return string(k) }

// Defines a set of transport Type
const (
	TypeGRPC Type = "grpc"
	TypeHTTP Type = "http"
)

var GrpcHostAddress string

// ITransport is transport context value interface.
type ITransport interface {
	// GetType transporter
	// grpc
	// http
	GetType() Type
	// GetEndpoint return server or client Endpoint
	// Server Transport: grpc://127.0.0.1:9000
	// Client Transport: discovery:///provider-demo
	GetEndpoint() string
	// GetOperation Service full method selector generated by protobuf
	// example: /helloworld.Greeter/SayHello
	GetOperation() string
	// GetReqHeader return transport request header
	// http: http.Header
	// grpc: metadata.MD
	GetReqHeader() header.IHeader
	// GetRspHeader return transport reply/response header
	// only valid for server transport
	// http: http.Header
	// grpc: metadata.MD
	GetRspHeader() header.IHeader
}

var _ ITransport = (*Transport)(nil)

// Transport is a gRPC transport.
type Transport struct {
	Type      Type
	Endpoint  string
	Operation string
	InHeader  header.IHeader
	OutHeader header.IHeader
	Request   *http.Request
}

// GetType returns the transport Type.
func (tr *Transport) GetType() Type {
	return tr.Type
}

// GetEndpoint returns the transport Endpoint.
func (tr *Transport) GetEndpoint() string {
	return tr.Endpoint
}

// GetOperation returns the transport Operation.
func (tr *Transport) GetOperation() string {
	return tr.Operation
}

// GetReqHeader returns the request header.
func (tr *Transport) GetReqHeader() header.IHeader {
	return tr.InHeader
}

// GetRspHeader returns the reply header.
func (tr *Transport) GetRspHeader() header.IHeader {
	return tr.OutHeader
}

// SetOperation sets the transport Operation.
func SetOperation(ctx context.Context, op string) {
	if tr, ok := FromContext(ctx); ok {
		if tr, ok := tr.(*Transport); ok {
			tr.Operation = op
		}
	}
}

type (
	transportKey struct{}
)

// ToContext returns a new Context that carries value.
func ToContext(ctx context.Context, tr ITransport) context.Context {
	return context.WithValue(ctx, transportKey{}, tr)
}

// FromContext returns the Transport value stored in ctx, if any.
func FromContext(ctx context.Context) (tr ITransport, ok bool) {
	tr, ok = ctx.Value(transportKey{}).(ITransport)
	return
}

func GetTransportFromHTTP(r *http.Request) ITransport {
	return &Transport{
		Type:      TypeHTTP,
		Endpoint:  r.URL.String(),
		Operation: r.URL.Path,
		InHeader:  (header.Header)(r.Header),
	}
}

func GetTransportFromGrpcClient(ctx context.Context, method string, cc *grpc.ClientConn, header header.IHeader) ITransport {
	return &Transport{
		Type:      TypeGRPC,
		Endpoint:  cc.Target(),
		Operation: method,
		InHeader:  header,
	}
}

func GetTransportFromGrpcServer(ctx context.Context, info *grpc.UnaryServerInfo, header header.IHeader) ITransport {
	return &Transport{
		Type:      TypeGRPC,
		Endpoint:  GrpcHostAddress,
		Operation: info.FullMethod,
		InHeader:  header,
	}
}
