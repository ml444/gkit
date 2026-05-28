package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"github.com/ml444/gkit/discovery"
)

type fakeEtcdClient struct {
	mu    sync.RWMutex
	store map[string]string
}

func newFakeEtcdClient() *fakeEtcdClient {
	return &fakeEtcdClient{store: make(map[string]string)}
}

func (f *fakeEtcdClient) Put(_ context.Context, key, val string, _ ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.store[key] = val
	return &clientv3.PutResponse{}, nil
}

func (f *fakeEtcdClient) Delete(_ context.Context, key string, _ ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.store, key)
	return &clientv3.DeleteResponse{}, nil
}

func (f *fakeEtcdClient) Get(_ context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	resp := &clientv3.GetResponse{}
	for k, v := range f.store {
		if k == key || strings.HasPrefix(k, key) {
			resp.Kvs = append(resp.Kvs, &mvccpb.KeyValue{Key: []byte(k), Value: []byte(v)})
		}
	}
	return resp, nil
}

func (f *fakeEtcdClient) Grant(_ context.Context, _ int64) (*clientv3.LeaseGrantResponse, error) {
	return &clientv3.LeaseGrantResponse{ID: 1}, nil
}

func (f *fakeEtcdClient) KeepAlive(_ context.Context, _ clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	ch := make(chan *clientv3.LeaseKeepAliveResponse)
	close(ch)
	return ch, nil
}

func (f *fakeEtcdClient) Revoke(_ context.Context, _ clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	return &clientv3.LeaseRevokeResponse{}, nil
}

func (f *fakeEtcdClient) Watch(_ context.Context, _ string, _ ...clientv3.OpOption) clientv3.WatchChan {
	ch := make(chan clientv3.WatchResponse)
	close(ch)
	return ch
}

func (f *fakeEtcdClient) Close() error { return nil }

func TestEtcdRegistry_RegisterGetDeregister(t *testing.T) {
	client := newFakeEtcdClient()
	reg, err := NewEtcdRegistryWithClient(client)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	ins := &discovery.ServiceInstance{
		ID: "1", Name: "svc", Address: "127.0.0.1", Port: 8080,
	}
	if err := reg.Register(ctx, ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	instances, err := reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 || instances[0].GetAddress() != "127.0.0.1" {
		t.Fatalf("unexpected instances: %+v", instances)
	}

	if err := reg.Deregister(ctx, ins); err != nil {
		t.Fatalf("deregister: %v", err)
	}
	_, err = reg.GetServiceInstances(ctx, "svc")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestEtcdRegistry_SkipInvalidJSON(t *testing.T) {
	client := newFakeEtcdClient()
	reg, err := NewEtcdRegistryWithClient(client)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	key := "/gkit/services/svc/bad"
	client.store[key] = "not-json"

	valid, _ := json.Marshal(&discovery.ServiceInstance{ID: "1", Name: "svc", Address: "10.0.0.1", Port: 80})
	client.store["/gkit/services/svc/1"] = string(valid)

	instances, err := reg.GetServiceInstances(context.Background(), "svc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 {
		t.Fatalf("expected 1 valid instance, got %d", len(instances))
	}
}

func TestEtcdRegistry_InvalidateCacheForKey(t *testing.T) {
	client := newFakeEtcdClient()
	reg, err := NewEtcdRegistryWithClient(client)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	valid, _ := json.Marshal(&discovery.ServiceInstance{ID: "1", Name: "svc", Address: "10.0.0.1", Port: 80})
	client.store["/gkit/services/svc/1"] = string(valid)

	if _, err := reg.GetServiceInstances(context.Background(), "svc"); err != nil {
		t.Fatalf("prime cache: %v", err)
	}

	reg.InvalidateCacheForKey("/gkit/services/svc/1")
	delete(client.store, "/gkit/services/svc/1")
	_, err = reg.GetServiceInstances(context.Background(), "svc")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after cache invalidation, got %v", err)
	}
}
