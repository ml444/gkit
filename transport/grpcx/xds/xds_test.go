package xds

import (
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestWithDialOptions(t *testing.T) {
	cfg := &clientConfig{}
	WithDialOptions(grpc.WithTransportCredentials(insecure.NewCredentials()))(cfg)
	if len(cfg.dialOpts) != 1 {
		t.Fatalf("dial opts = %d", len(cfg.dialOpts))
	}
}

func TestXDSBootstrapErrors(t *testing.T) {
	t.Setenv("GRPC_XDS_BOOTSTRAP", "/path/that/does/not/exist")
	if _, err := NewClient("xds:///listener"); err == nil {
		t.Fatal("expected xds client bootstrap error")
	}
	if _, err := NewGRPCServer(); err == nil {
		t.Fatal("expected xds server bootstrap error")
	}
}
