// Reference to kratos: https://github.com/go-kratos/kratos/blob/main/transport/transport.go
// I don’t design client Transport because I don’t think it’s necessary.

package transport

import (
	"context"
)

// ITransport is transport context value interface.
type ITransport interface {
	Kind() string
	// Endpoint return server endpoint
	// Server Transport: 127.0.0.1:9000
	// Client Transport: discovery:///provider-demo
	Endpoint() string
	// Method Service full method selector generated by protobuf
	// grpc example: /helloworld.Greeter/SayHello
	// http example: /path/{id}
	Method() string
	// InHeader return transport request header
	// http: http.Header
	// grpc: metadata.MD
	InHeader() MD
	// OutHeader return transport reply/response header
	// only valid for server transport
	// http: http.Header
	// grpc: metadata.MD
	OutHeader() MD
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
