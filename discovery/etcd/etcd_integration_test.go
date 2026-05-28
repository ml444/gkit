//go:build integration

package etcd

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/ml444/gkit/discovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
)

func startEmbeddedEtcd(t *testing.T) []string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	cfg := embed.NewConfig()
	cfg.Dir = t.TempDir()
	cfg.LogLevel = "error"
	cfg.LCUrls = []string{"http://" + addr}
	cfg.LPUrls = []string{"http://" + addr}

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatalf("start etcd: %v", err)
	}

	select {
	case <-e.Server.ReadyNotify():
	case <-time.After(10 * time.Second):
		e.Close()
		t.Fatal("etcd not ready")
	}

	t.Cleanup(func() {
		e.Close()
	})
	return []string{"http://" + addr}
}

func TestEtcdRegistry_Integration(t *testing.T) {
	endpoints := startEmbeddedEtcd(t)
	reg, err := NewEtcdRegistry(endpoints, WithTTL(5))
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	ins1 := &discovery.ServiceInstance{ID: "1", Name: "svc", Address: "10.0.0.1", Port: 8080}
	ins2 := &discovery.ServiceInstance{ID: "2", Name: "svc", Address: "10.0.0.2", Port: 8080}

	if err := reg.Register(ctx, ins1); err != nil {
		t.Fatalf("register ins1: %v", err)
	}
	if err := reg.Register(ctx, ins2); err != nil {
		t.Fatalf("register ins2: %v", err)
	}

	instances, err := reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}

	if err := reg.Deregister(ctx, ins1); err != nil {
		t.Fatalf("deregister: %v", err)
	}
	reg.serviceMap.Delete("svc")

	instances, err = reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get after deregister: %v", err)
	}
	if len(instances) != 1 || instances[0].GetID() != "2" {
		t.Fatalf("unexpected instances: %+v", instances)
	}

	client, err := clientv3.New(clientv3.Config{Endpoints: endpoints, DialTimeout: 5 * time.Second})
	if err != nil {
		t.Fatalf("client: %v", err)
	}
	defer client.Close()
	key := fmt.Sprintf("%s/svc/2", reg.basePath)
	resp, err := client.Get(ctx, key)
	if err != nil || len(resp.Kvs) != 1 {
		t.Fatalf("expected remaining key in etcd, err=%v kvs=%d", err, len(resp.Kvs))
	}
}
