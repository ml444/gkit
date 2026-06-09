package grpcx

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestServerStartAndStopErrorBranches(t *testing.T) {
	srv, err := NewServer(Network("bad-network"))
	if err != nil {
		t.Fatal(err)
	}
	if err := srv.Start(); err == nil {
		t.Fatal("expected start listen error")
	}

	blocking := &Server{iServer: blockingServer{}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	if err := blocking.Stop(ctx); err == nil {
		t.Fatal("expected forced stop error")
	}
}

func TestBuildServiceInstanceErrors(t *testing.T) {
	s := &Server{endpoint: "bad-endpoint"}
	if _, err := s.buildServiceInstance(); err == nil {
		t.Fatal("expected split endpoint error")
	}
	s = &Server{endpoint: "127.0.0.1:bad"}
	if _, err := s.buildServiceInstance(); err == nil {
		t.Fatal("expected parse port error")
	}
}

type blockingServer struct{}

func (blockingServer) RegisterService(*grpc.ServiceDesc, any) {}
func (blockingServer) Serve(net.Listener) error               { return nil }
func (blockingServer) Stop()                                  {}
func (blockingServer) GracefulStop()                          { time.Sleep(time.Second) }
func (blockingServer) GetServiceInfo() map[string]grpc.ServiceInfo {
	return nil
}
