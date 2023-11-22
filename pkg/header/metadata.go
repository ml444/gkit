package header

import (
	"context"
	"fmt"
)

type serverMetadataKey struct{}

// NewServerContext creates a new context with client md attached.
func NewServerContext(ctx context.Context, md Header) context.Context {
	return context.WithValue(ctx, serverMetadataKey{}, md)
}

// FromServerContext returns the server metadata in ctx if it exists.
func FromServerContext(ctx context.Context) (Header, bool) {
	md, ok := ctx.Value(serverMetadataKey{}).(Header)
	return md, ok
}

type clientMetadataKey struct{}

// NewClientContext creates a new context with client md attached.
func NewClientContext(ctx context.Context, md Header) context.Context {
	return context.WithValue(ctx, clientMetadataKey{}, md)
}

// FromClientContext returns the client metadata in ctx if it exists.
func FromClientContext(ctx context.Context) (Header, bool) {
	md, ok := ctx.Value(clientMetadataKey{}).(Header)
	return md, ok
}

// AppendToClientContext returns a new context with the provided kv merged
// with any existing metadata in the context.
func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
	if len(kv)%2 == 1 {
		panic(fmt.Sprintf("metadata: AppendToClientContext got an odd number of input pairs for metadata: %d", len(kv)))
	}
	md, _ := FromClientContext(ctx)
	md = md.Clone()
	for i := 0; i < len(kv); i += 2 {
		md.Set(kv[i], kv[i+1])
	}
	return NewClientContext(ctx, md)
}

// MergeToClientContext merge new metadata into ctx.
func MergeToClientContext(ctx context.Context, cmd Header) context.Context {
	md, _ := FromClientContext(ctx)
	md = md.Clone()
	for k, v := range cmd {
		md[k] = v
	}
	return NewClientContext(ctx, md)
}
