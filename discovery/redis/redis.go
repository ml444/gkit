package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"


	"github.com/redis/go-redis/v9"
	"github.com/ml444/gkit/discovery"
)

var _ discovery.ServiceRegistry = (*RedisRegistry)(nil)

type RedisRegistry struct {
	client     redis.UniversalClient
	addr       string
	serviceTTL int64
	prefix     string
	serviceMap sync.Map
	closeCh    chan struct{}
	closeOnce  sync.Once
	wg         sync.WaitGroup
}

type RedisRegistryOption func(*RedisRegistry)

func WithTTL(ttl int64) RedisRegistryOption {
	return func(r *RedisRegistry) {
		r.serviceTTL = ttl
	}
}

func WithPrefix(prefix string) RedisRegistryOption {
	return func(r *RedisRegistry) {
		r.prefix = prefix
	}
}

func WithAddr(addr string) RedisRegistryOption {
	return func(r *RedisRegistry) {
		r.addr = addr
	}
}

func WithClient(client redis.UniversalClient) RedisRegistryOption {
	return func(r *RedisRegistry) {
		r.client = client
	}
}

func newRedisRegistry(options ...RedisRegistryOption) (*RedisRegistry, error) {
	registry := &RedisRegistry{
		serviceTTL: 60,
		prefix:     "gkit:services",
		addr:       "localhost:6379",
		closeCh:    make(chan struct{}),
	}

	for _, option := range options {
		option(registry)
	}

	if registry.client == nil {
		registry.client = redis.NewClient(&redis.Options{Addr: registry.addr})
	}

	ctx := context.Background()
	if err := registry.client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	registry.startWatching()
	return registry, nil
}

func NewRedisRegistry(options ...RedisRegistryOption) (*RedisRegistry, error) {
	return newRedisRegistry(options...)
}

func NewRedisRegistryWithClient(client redis.UniversalClient, options ...RedisRegistryOption) (*RedisRegistry, error) {
	return newRedisRegistry(append([]RedisRegistryOption{WithClient(client)}, options...)...)
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

// getServiceKey returns the redis key for a service
func (r *RedisRegistry) getServiceKey(serviceName string) string {
	return fmt.Sprintf("%s:%s", r.prefix, serviceName)
}

// getInstanceKey returns the redis key for a service instance
func (r *RedisRegistry) getInstanceKey(serviceName, instanceID string) string {
	return fmt.Sprintf("%s:%s:%s", r.prefix, serviceName, instanceID)
}

// Register registers a service instance
func (r *RedisRegistry) Register(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	instanceData, err := marshalInstance(instance)
	if err != nil {
		return fmt.Errorf("failed to marshal instance: %w", err)
	}

	// Store instance with TTL
	instanceKey := r.getInstanceKey(instance.GetName(), instance.GetID())
	err = r.client.Set(ctx, instanceKey, instanceData, time.Duration(r.serviceTTL)*time.Second).Err()
	if err != nil {
		return fmt.Errorf("failed to store instance: %v", err)
	}

	// Add instance ID to service set
	serviceKey := r.getServiceKey(instance.GetName())
	err = r.client.SAdd(ctx, serviceKey, instance.GetID()).Err()
	if err != nil {
		// Clean up if adding to set fails
		r.client.Del(ctx, instanceKey)
		return fmt.Errorf("failed to add instance to service set: %v", err)
	}

	// Update service map
	services, _ := r.serviceMap.LoadOrStore(instance.GetName(), make([]discovery.ServiceInstancer, 0))

	// Create a copy of the slice to avoid race conditions
	instances := services.([]discovery.ServiceInstancer)
	newInstances := make([]discovery.ServiceInstancer, len(instances)+1)
	copy(newInstances, instances)
	newInstances[len(instances)] = instance

	r.serviceMap.Store(instance.GetName(), newInstances)

	return nil
}

// Deregister deregisters a service instance
func (r *RedisRegistry) Deregister(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	// Remove instance from service set
	serviceKey := r.getServiceKey(instance.GetName())
	removed, err := r.client.SRem(ctx, serviceKey, instance.GetID()).Result()
	if err != nil {
		return fmt.Errorf("failed to remove instance from service set: %v", err)
	}

	// If instance was not in the set, return not found
	if removed == 0 {
		return discovery.ErrNotFound
	}

	// Delete instance data
	instanceKey := r.getInstanceKey(instance.GetName(), instance.GetID())
	r.client.Del(ctx, instanceKey)

	// Check if service set is empty
	count, err := r.client.SCard(ctx, serviceKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check service set count: %v", err)
	}

	// Delete service set if empty
	if count == 0 {
		r.client.Del(ctx, serviceKey)
	}

	// Update service map
	services, ok := r.serviceMap.Load(instance.GetName())
	if !ok {
		return discovery.ErrNotFound
	}

	instances := services.([]discovery.ServiceInstancer)
	newInstances := make([]discovery.ServiceInstancer, 0, len(instances))

	for _, ins := range instances {
		if ins.GetID() != instance.GetID() {
			newInstances = append(newInstances, ins)
		}
	}

	if len(newInstances) == 0 {
		r.serviceMap.Delete(instance.GetName())
	} else {
		r.serviceMap.Store(instance.GetName(), newInstances)
	}

	return nil
}

// GetServiceInstances gets all instances of a service
func (r *RedisRegistry) GetServiceInstances(ctx context.Context, serviceName string) ([]discovery.ServiceInstancer, error) {
	// Try to get from local cache first
	services, ok := r.serviceMap.Load(serviceName)
	if ok {
		return copyInstances(services.([]discovery.ServiceInstancer)), nil
	}

	// Get from Redis
	serviceKey := r.getServiceKey(serviceName)
	instanceIDs, err := r.client.SMembers(ctx, serviceKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get service instances: %v", err)
	}

	// If no instances found
	if len(instanceIDs) == 0 {
		return nil, discovery.ErrNotFound
	}

	instances := make([]discovery.ServiceInstancer, 0, len(instanceIDs))

	// Get each instance data
	for _, instanceID := range instanceIDs {
		instanceKey := r.getInstanceKey(serviceName, instanceID)
		instanceData, err := r.client.Get(ctx, instanceKey).Bytes()
		if err != nil {
			// Skip if instance data not found or expired
			if err == redis.Nil {
				// Clean up stale instance ID
				r.client.SRem(ctx, serviceKey, instanceID)
			}
			continue
		}

		var instance discovery.ServiceInstance
		if err := json.Unmarshal(instanceData, &instance); err != nil {
			// Skip if we can't unmarshal the data
			continue
		}

		instances = append(instances, &instance)
	}

	// Update local cache
	if len(instances) == 0 {
		return nil, discovery.ErrNotFound
	}

	r.serviceMap.Store(serviceName, instances)
	return copyInstances(instances), nil
}

// startWatching starts watching for service changes
func (r *RedisRegistry) startWatching() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		// Periodically refresh services to clean up expired instances
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
				// Refresh synchronously within this tracked goroutine so we do
				// not spawn an unbounded number of untracked goroutines.
				r.refreshAllServices()
			}
		}
	}()
}

// refreshAllServices refreshes all services in the local cache
func (r *RedisRegistry) refreshAllServices() {
	// In Redis, we don't have a direct way to list all service keys
	// Instead, we'll refresh services that are already in our local cache.
	r.serviceMap.Range(func(key, _ interface{}) bool {
		serviceName := key.(string)
		// Drop the cached entry first so GetServiceInstances actually re-reads
		// from Redis; otherwise the refresh is a no-op (cache hit) and expired
		// instances are never evicted.
		r.serviceMap.Delete(serviceName)
		_, _ = r.GetServiceInstances(context.Background(), serviceName)
		return true
	})
}

// Close closes the registry
func (r *RedisRegistry) Close() error {
	var err error
	r.closeOnce.Do(func() {
		// Signal goroutines to stop
		close(r.closeCh)
		// Wait for goroutines to finish
		r.wg.Wait()
		// Close the client connection
		err = r.client.Close()
	})
	return err
}