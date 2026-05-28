package grpcx

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/ml444/gkit/discovery"
)

func TestClient_DiscoveryEndpoint(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	_, portStr, _ := net.SplitHostPort(ln.Addr().String())
	port, _ := strconv.Atoi(portStr)

	srv := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	healthpb.RegisterHealthServer(srv, health.NewServer())
	go func() { _ = srv.Serve(ln) }()
	defer srv.GracefulStop()

	reg := discovery.NewDefaultRegistry()
	dc := discovery.NewDiscoveryClient(reg, discovery.WithCacheTTL(time.Minute))
	if err := reg.Register(context.Background(), &discovery.ServiceInstance{
		ID: "ins-1", Name: "svc", Address: "127.0.0.1", Port: port,
	}); err != nil {
		t.Fatal(err)
	}

	client, err := NewClient(
		WithEndpoint("discovery:///svc"),
		WithDiscovery(dc, ""),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	hc := healthpb.NewHealthClient(client.Conn())
	_, err = hc.Check(ctx, &healthpb.HealthCheckRequest{Service: ""})
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
}
