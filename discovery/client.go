package discovery

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// DiscoveryClient is a client for service discovery
// 服务发现客户端

type DiscoveryClient struct {
	registry     ServiceRegistry
	loadBalancer LoadBalancer
	cache        sync.Map // key: service name, value: cached instances with timestamp
	cacheTTL     time.Duration
	fetchGroup   singleflight.Group
}

// DiscoveryClientOption is option for DiscoveryClient
// 服务发现客户端配置选项

type DiscoveryClientOption func(*DiscoveryClient)

// WithLoadBalancer sets the load balancer for the client
// 设置客户端的负载均衡器

func WithLoadBalancer(lb LoadBalancer) DiscoveryClientOption {
	return func(c *DiscoveryClient) {
		c.loadBalancer = lb
	}
}

// WithCacheTTL sets the cache TTL for service instances
// 设置服务实例缓存的TTL

func WithCacheTTL(ttl time.Duration) DiscoveryClientOption {
	return func(c *DiscoveryClient) {
		c.cacheTTL = ttl
	}
}

// WithRegistry sets the service registry for the client
// 设置客户端的服务注册中心

func WithRegistry(registry ServiceRegistry) DiscoveryClientOption {
	return func(c *DiscoveryClient) {
		c.registry = registry
	}
}

// WithRefreshInterval sets the refresh interval for service instances
// 设置服务实例刷新间隔（与WithCacheTTL功能相同，为兼容保留）

func WithRefreshInterval(interval time.Duration) DiscoveryClientOption {
	return func(c *DiscoveryClient) {
		c.cacheTTL = interval
	}
}

// cachedInstances represents cached service instances with timestamp
// 带时间戳的缓存服务实例

type cachedInstances struct {
	instances []ServiceInstancer
	timestamp time.Time
}

// NewDiscoveryClient creates a new DiscoveryClient
// 创建一个新的服务发现客户端
func NewDiscoveryClient(registry ServiceRegistry, options ...DiscoveryClientOption) *DiscoveryClient {
	client := &DiscoveryClient{
		registry:     registry,
		loadBalancer: NewRoundRobinLoadBalancer(),
		cacheTTL:     30 * time.Second,
	}

	for _, option := range options {
		option(client)
	}

	return client
}

func copyInstances(instances []ServiceInstancer) []ServiceInstancer {
	result := make([]ServiceInstancer, len(instances))
	copy(result, instances)
	return result
}

func (c *DiscoveryClient) fetchInstances(ctx context.Context, serviceName string) ([]ServiceInstancer, error) {
	value, err, _ := c.fetchGroup.Do(serviceName, func() (any, error) {
		instances, err := c.registry.GetServiceInstances(ctx, serviceName)
		if err != nil {
			return nil, fmt.Errorf("failed to get service instances: %w", err)
		}
		if len(instances) == 0 {
			return nil, ErrNotFound
		}
		c.storeCache(serviceName, instances)
		return instances, nil
	})
	if err != nil {
		return nil, err
	}
	return value.([]ServiceInstancer), nil
}

func (c *DiscoveryClient) storeCache(serviceName string, instances []ServiceInstancer) {
	if c.cacheTTL <= 0 {
		return
	}
	c.cache.Store(serviceName, &cachedInstances{
		instances: copyInstances(instances),
		timestamp: time.Now(),
	})
}

// GetServiceInstance gets a service instance using load balancing
// 使用负载均衡获取一个服务实例
func (c *DiscoveryClient) GetServiceInstance(ctx context.Context, serviceName string) (ServiceInstancer, error) {
	instances, err := c.getInstancesFromCache(serviceName)
	if err == nil && len(instances) > 0 {
		return c.loadBalancer.Select(ctx, instances)
	}

	instances, err = c.fetchInstances(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	return c.loadBalancer.Select(ctx, instances)
}

// GetAllInstances gets all instances of a service
// 获取服务的所有实例
func (c *DiscoveryClient) GetAllInstances(ctx context.Context, serviceName string) ([]ServiceInstancer, error) {
	instances, err := c.getInstancesFromCache(serviceName)
	if err == nil && len(instances) > 0 {
		return copyInstances(instances), nil
	}

	instances, err = c.fetchInstances(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	return copyInstances(instances), nil
}

// UpdateLoadBalancerStatus updates the load balancer status
// 更新负载均衡器状态
func (c *DiscoveryClient) UpdateLoadBalancerStatus(ctx context.Context, instance ServiceInstancer, success bool) {
	c.loadBalancer.Update(ctx, instance, success)
}

func (c *DiscoveryClient) getInstancesFromCache(serviceName string) ([]ServiceInstancer, error) {
	if c.cacheTTL <= 0 {
		return nil, ErrNotFound
	}

	value, ok := c.cache.Load(serviceName)
	if !ok {
		return nil, ErrNotFound
	}

	cached := value.(*cachedInstances)
	if time.Since(cached.timestamp) > c.cacheTTL {
		c.cache.Delete(serviceName)
		return nil, ErrNotFound
	}

	return cached.instances, nil
}

// RefreshCache refreshes the cache for a specific service
// 刷新特定服务的缓存
func (c *DiscoveryClient) RefreshCache(ctx context.Context, serviceName string) error {
	instances, err := c.registry.GetServiceInstances(ctx, serviceName)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.cache.Delete(serviceName)
			return nil
		}
		return fmt.Errorf("failed to get service instances: %w", err)
	}

	if len(instances) == 0 {
		c.cache.Delete(serviceName)
		return nil
	}

	c.storeCache(serviceName, instances)
	return nil
}

// ClearCache clears the cache
// 清除缓存
func (c *DiscoveryClient) ClearCache() {
	c.cache = sync.Map{}
}

// Close closes the client and the underlying registry
// 关闭客户端和底层注册中心
func (c *DiscoveryClient) Close() error {
	return c.registry.Close()
}
