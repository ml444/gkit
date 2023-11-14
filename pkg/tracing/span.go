package tracing

import (
	"context"
	"net"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

func SetClientSpan(ctx context.Context, span trace.Span, m interface{}) {
	var (
		attrs     []attribute.KeyValue
		remote    string
		operation string
		rpcKind   string
	)
	attrs = append(attrs, semconv.RPCSystemKey.String(rpcKind))
	_, mAttrs := parseFullMethod(operation)
	attrs = append(attrs, mAttrs...)
	if remote != "" {
		attrs = append(attrs, peerAttr(remote)...)
	}
	if p, ok := m.(proto.Message); ok {
		attrs = append(attrs, attribute.Key("send_msg.size").Int(proto.Size(p)))
	}

	span.SetAttributes(attrs...)

}

func SetServerSpan(ctx context.Context, span trace.Span, m interface{}) {
	var (
		attrs     []attribute.KeyValue
		remote    string
		operation string
		rpcKind   string
	)
	attrs = append(attrs, semconv.RPCSystemKey.String(rpcKind))
	_, mAttrs := parseFullMethod(operation)
	attrs = append(attrs, mAttrs...)
	attrs = append(attrs, peerAttr(remote)...)
	if p, ok := m.(proto.Message); ok {
		attrs = append(attrs, attribute.Key("recv_msg.size").Int(proto.Size(p)))
	}
	//if md, ok := metadata.FromServerContext(ctx); ok {
	//	attrs = append(attrs, semconv.PeerServiceKey.String(md.Get(serviceHeader)))
	//}

	span.SetAttributes(attrs...)

}

// parseFullMethod returns a span name following the OpenTelemetry semantic
// conventions as well as all applicable span attribute.KeyValue attributes based
// on a gRPC's FullMethod.
func parseFullMethod(fullMethod string) (string, []attribute.KeyValue) {
	name := strings.TrimLeft(fullMethod, "/")
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 { //nolint:gomnd
		// Invalid format, does not follow `/package.service/method`.
		return name, []attribute.KeyValue{attribute.Key("rpc.operation").String(fullMethod)}
	}

	var attrs []attribute.KeyValue
	if service := parts[0]; service != "" {
		attrs = append(attrs, semconv.RPCServiceKey.String(service))
	}
	if method := parts[1]; method != "" {
		attrs = append(attrs, semconv.RPCMethodKey.String(method))
	}
	return name, attrs
}

// peerAttr returns attributes about the peer address.
func peerAttr(addr string) []attribute.KeyValue {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return []attribute.KeyValue(nil)
	}

	if host == "" {
		host = "127.0.0.1"
	}

	return []attribute.KeyValue{
		semconv.NetPeerIPKey.String(host),
		semconv.NetPeerPortKey.String(port),
	}
}

func parseTarget(endpoint string) (address string, err error) {
	var u *url.URL
	u, err = url.Parse(endpoint)
	if err != nil {
		if u, err = url.Parse("http://" + endpoint); err != nil {
			return "", err
		}
		return u.Host, nil
	}
	if len(u.Path) > 1 {
		return u.Path[1:], nil
	}
	return endpoint, nil
}
