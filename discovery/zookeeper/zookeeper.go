package zookeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/ml444/gkit/discovery"
)

var _ discovery.ServiceRegistry = (*ZookeeperRegistry)(nil)

type zkConn interface {
	Exists(path string) (bool, *zk.Stat, error)
	Create(path string, data []byte, flags int32, acl []zk.ACL) (string, error)
	Delete(path string, version int32) error
	Get(path string) ([]byte, *zk.Stat, error)
	Children(path string) ([]string, *zk.Stat, error)
	Close()
}

type ZookeeperRegistry struct {
	client     zkConn
	serviceTTL int64
	basePath   string
	mu         sync.RWMutex
	services   map[string][]discovery.ServiceInstancer
	closeCh    chan struct{}
	wg         sync.WaitGroup
}

type ZookeeperRegistryOption func(*ZookeeperRegistry)

func WithTTL(ttl int64) ZookeeperRegistryOption {
	return func(r *ZookeeperRegistry) {
		r.serviceTTL = ttl
	}
}

func WithBasePath(basePath string) ZookeeperRegistryOption {
	return func(r *ZookeeperRegistry) {
		r.basePath = basePath
	}
}

func NewZookeeperRegistry(endpoints []string, options ...ZookeeperRegistryOption) (*ZookeeperRegistry, error) {
	conn, _, err := zk.Connect(endpoints, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to zookeeper: %w", err)
	}
	return NewZookeeperRegistryWithConn(conn, options...)
}

func NewZookeeperRegistryWithConn(conn zkConn, options ...ZookeeperRegistryOption) (*ZookeeperRegistry, error) {
	registry := &ZookeeperRegistry{
		client:     conn,
		serviceTTL: 60,
		basePath:   "/gkit/services",
		services:   make(map[string][]discovery.ServiceInstancer),
		closeCh:    make(chan struct{}),
	}

	for _, option := range options {
		option(registry)
	}

	if err := registry.ensurePath(registry.basePath); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create base path: %w", err)
	}

	registry.startWatching()
	return registry, nil
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

func (r *ZookeeperRegistry) ensurePath(path string) error {
	exists, _, err := r.client.Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		parts := strings.Split(path, "/")
		current := ""
		for i := 1; i < len(parts); i++ {
			current += "/" + parts[i]
			exists, _, err := r.client.Exists(current)
			if err != nil {
				return err
			}
			if !exists {
				if _, err := r.client.Create(current, nil, 0, zk.WorldACL(zk.PermAll)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (r *ZookeeperRegistry) instancePath(instance discovery.ServiceInstancer) string {
	return fmt.Sprintf("%s/%s/%s", r.basePath, instance.GetName(), instance.GetID())
}

func (r *ZookeeperRegistry) upsertLocal(instance discovery.ServiceInstancer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[instance.GetName()]
	for i, ins := range instances {
		if ins.GetID() == instance.GetID() {
			instances[i] = instance
			r.services[instance.GetName()] = instances
			return
		}
	}
	r.services[instance.GetName()] = append(instances, instance)
}

func (r *ZookeeperRegistry) removeLocal(serviceName, instanceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances, ok := r.services[serviceName]
	if !ok {
		return discovery.ErrNotFound
	}

	newInstances := make([]discovery.ServiceInstancer, 0, len(instances))
	for _, ins := range instances {
		if ins.GetID() != instanceID {
			newInstances = append(newInstances, ins)
		}
	}

	if len(newInstances) == 0 {
		delete(r.services, serviceName)
	} else {
		r.services[serviceName] = newInstances
	}
	return nil
}

func (r *ZookeeperRegistry) Register(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	servicePath := fmt.Sprintf("%s/%s", r.basePath, instance.GetName())
	if err := r.ensurePath(servicePath); err != nil {
		return err
	}

	instanceData, err := marshalInstance(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	if _, err = r.client.Create(r.instancePath(instance), instanceData, zk.FlagEphemeral, zk.WorldACL(zk.PermAll)); err != nil {
		return fmt.Errorf("failed to create instance node: %w", err)
	}

	r.upsertLocal(instance)
	return nil
}

func (r *ZookeeperRegistry) Deregister(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	if err := r.client.Delete(r.instancePath(instance), -1); err != nil && err != zk.ErrNoNode {
		return fmt.Errorf("failed to delete instance node: %w", err)
	}

	return r.removeLocal(instance.GetName(), instance.GetID())
}

func (r *ZookeeperRegistry) GetServiceInstances(ctx context.Context, serviceName string) ([]discovery.ServiceInstancer, error) {
	r.mu.RLock()
	if instances, ok := r.services[serviceName]; ok {
		result := make([]discovery.ServiceInstancer, len(instances))
		copy(result, instances)
		r.mu.RUnlock()
		return result, nil
	}
	r.mu.RUnlock()

	servicePath := fmt.Sprintf("%s/%s", r.basePath, serviceName)
	children, _, err := r.client.Children(servicePath)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, discovery.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get children: %w", err)
	}
	if len(children) == 0 {
		return nil, discovery.ErrNotFound
	}

	instances := make([]discovery.ServiceInstancer, 0, len(children))
	for _, child := range children {
		data, _, err := r.client.Get(fmt.Sprintf("%s/%s", servicePath, child))
		if err != nil {
			continue
		}
		var instance discovery.ServiceInstance
		if err := json.Unmarshal(data, &instance); err != nil {
			continue
		}
		instances = append(instances, &instance)
	}

	if len(instances) == 0 {
		return nil, discovery.ErrNotFound
	}

	r.mu.Lock()
	r.services[serviceName] = instances
	r.mu.Unlock()

	result := make([]discovery.ServiceInstancer, len(instances))
	copy(result, instances)
	return result, nil
}

func (r *ZookeeperRegistry) startWatching() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		interval := r.serviceTTL / 2
		if interval <= 0 {
			interval = 1
		}
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.closeCh:
				return
			case <-ticker.C:
				go r.refreshAllServices()
			}
		}
	}()
}

func (r *ZookeeperRegistry) refreshAllServices() {
	serviceNames, _, err := r.client.Children(r.basePath)
	if err != nil {
		return
	}
	for _, serviceName := range serviceNames {
		r.mu.Lock()
		delete(r.services, serviceName)
		r.mu.Unlock()
		_, _ = r.GetServiceInstances(context.Background(), serviceName)
	}
}

func (r *ZookeeperRegistry) Close() error {
	close(r.closeCh)
	r.wg.Wait()
	r.client.Close()
	return nil
}
