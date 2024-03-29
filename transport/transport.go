package transport

import (
	"context"

	"github.com/ml444/gkit/pkg/header"
)

type Kind uint8

func (k Kind) String() string {
	switch k {
	case KindGRPC:
		return "GRPC"
	case KindHTTP:
		return "HTTP"
	default:
		return ""
	}
}

// Defines a set of transport Type
const (
	KindGRPC Kind = 1
	KindHTTP Kind = 2
)

// ITransport is transport context value interface.
type ITransport interface {
	GetKind() Kind
	// GetEndpoint return server or client endpoint
	// Server Transport: grpc://127.0.0.1:9000
	// Client Transport: discovery:///provider-demo
	GetEndpoint() string
	// GetOperation Service full method selector generated by protobuf
	// example: /helloworld.Greeter/SayHello
	GetOperation() string
	// GetInHeader return transport request header
	// http: http.Header
	// grpc: metadata.MD
	GetInHeader() header.IHeader
	// GetOutHeader return transport reply/response header
	// only valid for server transport
	// http: http.Header
	// grpc: metadata.MD
	GetOutHeader() header.IHeader
}

type BaseTransport struct {
	Endpoint  string
	Operation string
	InHeader  header.IHeader
	OutHeader header.IHeader
}

func (tr *BaseTransport) GetEndpoint() string {
	return tr.Endpoint
}

func (tr *BaseTransport) GetOperation() string {
	return tr.Operation
}

func (tr *BaseTransport) GetInHeader() header.IHeader {
	return tr.InHeader
}

func (tr *BaseTransport) GetOutHeader() header.IHeader {
	return tr.OutHeader
}

type transportKey struct{}

// ToContext returns a new Context that carries value.
func ToContext(ctx context.Context, tr ITransport) context.Context {
	return context.WithValue(ctx, transportKey{}, tr)
}

// FromContext returns the Transport value stored in ctx, if any.
func FromContext(ctx context.Context) (tr ITransport, ok bool) {
	tr, ok = ctx.Value(transportKey{}).(ITransport)
	return
}
