package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	discovery "github.com/ml444/gkit/discovery"
)

var _ discovery.ServiceRegistry = (*EtcdRegistry)(nil)

type etcdClient interface {
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Grant(ctx context.Context, ttl int64) (*clientv3.LeaseGrantResponse, error)
	KeepAlive(ctx context.Context, id clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error)
	Revoke(ctx context.Context, id clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error)
	Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
	Close() error
}

type EtcdRegistry struct {
	client     etcdClient
	serviceTTL int64
	basePath   string
	serviceMap sync.Map
	leases     sync.Map
	closeCh    chan struct{}
	closeOnce  sync.Once
}

type EtcdRegistryOption func(*EtcdRegistry)

func WithTTL(ttl int64) EtcdRegistryOption {
	return func(r *EtcdRegistry) {
		r.serviceTTL = ttl
	}
}

func WithBasePath(basePath string) EtcdRegistryOption {
	return func(r *EtcdRegistry) {
		r.basePath = basePath
	}
}

func NewEtcdRegistry(endpoints []string, options ...EtcdRegistryOption) (*EtcdRegistry, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}
	return NewEtcdRegistryWithClient(client, options...)
}

func NewEtcdRegistryWithClient(client etcdClient, options ...EtcdRegistryOption) (*EtcdRegistry, error) {
	registry := &EtcdRegistry{
		client:     client,
		serviceTTL: 60,
		basePath:   "/gkit/services",
		closeCh:    make(chan struct{}),
	}

	for _, option := range options {
		option(registry)
	}

	registry.startWatching()
	return registry, nil
}

func instanceKey(basePath string, instance discovery.ServiceInstancer) string {
	return fmt.Sprintf("%s/%s/%s", basePath, instance.GetName(), instance.GetID())
}

func marshalInstance(instance discovery.ServiceInstancer) ([]byte, error) {
	if si, ok := instance.(*discovery.ServiceInstance); ok {
		return json.Marshal(si)
	}
	return json.Marshal(&discovery.ServiceInstance{
		ID:          instance.GetID(),
		Name:        instance.GetName(),
		Version:     instance.GetVersion(),
		Address:     instance.GetAddress(),
		Port:        instance.GetPort(),
		Metadata:    instance.GetMetadata(),
		HealthCheck: instance.GetHealthCheck(),
	})
}

func copyInstances(instances []discovery.ServiceInstancer) []discovery.ServiceInstancer {
	result := make([]discovery.ServiceInstancer, len(instances))
	copy(result, instances)
	return result
}

func (r *EtcdRegistry) grantAndKeepAlive(ctx context.Context) (clientv3.LeaseID, error) {
	resp, err := r.client.Grant(ctx, r.serviceTTL)
	if err != nil {
		return 0, fmt.Errorf("failed to create lease: %w", err)
	}

	ch, err := r.client.KeepAlive(ctx, resp.ID)
	if err != nil {
		_, _ = r.client.Revoke(ctx, resp.ID)
		return 0, fmt.Errorf("failed to keepalive: %w", err)
	}

	go func(leaseID clientv3.LeaseID) {
		for {
			select {
			case _, ok := <-ch:
				if !ok {
					return
				}
			case <-r.closeCh:
				return
			}
		}
	}(resp.ID)

	return resp.ID, nil
}

func (r *EtcdRegistry) Register(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	key := instanceKey(r.basePath, instance)
	value, err := marshalInstance(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	leaseID, err := r.grantAndKeepAlive(ctx)
	if err != nil {
		return err
	}

	_, err = r.client.Put(ctx, key, string(value), clientv3.WithLease(leaseID))
	if err != nil {
		_, _ = r.client.Revoke(ctx, leaseID)
		return fmt.Errorf("failed to put instance into etcd: %w", err)
	}

	r.leases.Store(key, leaseID)
	r.serviceMap.Delete(instance.GetName())
	return nil
}

func (r *EtcdRegistry) Deregister(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	key := instanceKey(r.basePath, instance)
	if leaseVal, ok := r.leases.Load(key); ok {
		if leaseID, ok := leaseVal.(clientv3.LeaseID); ok {
			_, _ = r.client.Revoke(ctx, leaseID)
		}
		r.leases.Delete(key)
	}

	if _, err := r.client.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete instance from etcd: %w", err)
	}

	r.serviceMap.Delete(instance.GetName())
	return nil
}

func (r *EtcdRegistry) GetServiceInstances(ctx context.Context, serviceName string) ([]discovery.ServiceInstancer, error) {
	if instances, ok := r.serviceMap.Load(serviceName); ok {
		return copyInstances(instances.([]discovery.ServiceInstancer)), nil
	}

	prefix := fmt.Sprintf("%s/%s/", r.basePath, serviceName)
	resp, err := r.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("failed to get service instances: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, discovery.ErrNotFound
	}

	instances := make([]discovery.ServiceInstancer, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var instance discovery.ServiceInstance
		if err := json.Unmarshal(kv.Value, &instance); err != nil {
			continue
		}
		instances = append(instances, &instance)
	}

	r.serviceMap.Store(serviceName, instances)
	return copyInstances(instances), nil
}

func (r *EtcdRegistry) Close() error {
	r.closeOnce.Do(func() {
		close(r.closeCh)
	})

	r.leases.Range(func(key, value any) bool {
		if leaseID, ok := value.(clientv3.LeaseID); ok {
			_, _ = r.client.Revoke(context.Background(), leaseID)
		}
		r.leases.Delete(key)
		return true
	})

	return r.client.Close()
}

func (r *EtcdRegistry) invalidateCacheForKey(key string) {
	keyParts := strings.Split(key, "/")
	if len(keyParts) >= 2 {
		serviceName := keyParts[len(keyParts)-2]
		r.serviceMap.Delete(serviceName)
	}
}

func (r *EtcdRegistry) startWatching() {
	watchChan := r.client.Watch(context.Background(), r.basePath, clientv3.WithPrefix())

	go func() {
		for {
			select {
			case <-r.closeCh:
				return
			case resp, ok := <-watchChan:
				if !ok {
					watchChan = r.client.Watch(context.Background(), r.basePath, clientv3.WithPrefix())
					continue
				}
				for _, event := range resp.Events {
					r.invalidateCacheForKey(string(event.Kv.Key))
				}
			}
		}
	}()
}

// InvalidateCacheForKey exposes cache invalidation for tests.
func (r *EtcdRegistry) InvalidateCacheForKey(key string) {
	r.invalidateCacheForKey(key)
}
