package discovery

import (
	"context"
	"errors"
	"sync"
)


type ServiceInstancer interface {
	// GetID returns the unique identifier of the service instance
	// 获取服务实例的唯一标识符
	GetID() string

	// GetName returns the name of the service
	// 获取服务名称
	GetName() string

	// GetVersion returns the version of the service
	// 获取服务版本
	GetVersion() string

	// GetAddress returns the address of the service
	// 获取服务地址
	GetAddress() string

	// GetPort returns the port of the service
	// 获取服务端口
	GetPort() int

	// GetMetadata returns the metadata of the service instance
	// 获取服务实例的元数据
	GetMetadata() map[string]string

	// GetHealthCheck returns the health check URL of the service instance
	// 获取服务实例的健康检查URL
	GetHealthCheck() string
}

// ServiceInstance represents a service instance
// 服务实例
type ServiceInstance struct {
	ID          string            `json:"id"`           // Unique identifier of the service instance
	Name        string            `json:"name"`         // Service name
	Version     string            `json:"version"`      // Service version
	Address     string            `json:"address"`      // Service address
	Port        int               `json:"port"`         // Service port
	Metadata    map[string]string `json:"metadata"`     // Additional metadata
	HealthCheck string            `json:"health_check"` // Health check URL
}
func (s *ServiceInstance) GetID() string {
	return s.ID
}
func (s *ServiceInstance) GetName() string {
	return s.Name
}
func (s *ServiceInstance) GetVersion() string {
	return s.Version
}
func (s *ServiceInstance) GetAddress() string {
	return s.Address
}
func (s *ServiceInstance) GetPort() int {
	return s.Port
}
func (s *ServiceInstance) GetMetadata() map[string]string {
	return s.Metadata
}
func (s *ServiceInstance) GetHealthCheck() string {
	return s.HealthCheck
}

// ServiceRegistry is the interface for service registration and discovery
// 服务注册发现接口
type ServiceRegistry interface {
	// Register registers a service instance
	// 注册服务实例
	Register(ctx context.Context, instance ServiceInstancer) error

	// Deregister deregisters a service instance
	// 注销服务实例
	Deregister(ctx context.Context, instance ServiceInstancer) error

	// GetServiceInstances gets all instances of a service
	// 获取服务的所有实例
	GetServiceInstances(ctx context.Context, serviceName string) ([]ServiceInstancer, error)

	// Close closes the registry
	// 关闭注册中心连接
	Close() error
}

// ErrNotFound is returned when a service is not found
var ErrNotFound = errors.New("service not found")

// DefaultRegistry is an in-memory service registry implementation
// 默认的内存服务注册实现
type DefaultRegistry struct {
	mu       sync.RWMutex
	services map[string][]ServiceInstancer
}

// NewDefaultRegistry creates a new DefaultRegistry
func NewDefaultRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		services: make(map[string][]ServiceInstancer),
	}
}

// Register registers a service instance
func (r *DefaultRegistry) Register(ctx context.Context, instance ServiceInstancer) error {
	if instance == nil {
		return errors.New("instance cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[instance.GetName()]
	for i, ins := range instances {
		if ins.GetID() == instance.GetID() {
			instances[i] = instance
			r.services[instance.GetName()] = instances
			return nil
		}
	}

	r.services[instance.GetName()] = append(instances, instance)
	return nil
}

// Deregister deregisters a service instance
func (r *DefaultRegistry) Deregister(ctx context.Context, instance ServiceInstancer) error {
	if instance == nil {
		return errors.New("instance cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	instances, ok := r.services[instance.GetName()]
	if !ok {
		return ErrNotFound
	}

	newInstances := make([]ServiceInstancer, 0, len(instances))
	for _, ins := range instances {
		if ins.GetID() != instance.GetID() {
			newInstances = append(newInstances, ins)
		}
	}

	if len(newInstances) == 0 {
		delete(r.services, instance.GetName())
	} else {
		r.services[instance.GetName()] = newInstances
	}

	return nil
}

// GetServiceInstances gets all instances of a service
func (r *DefaultRegistry) GetServiceInstances(ctx context.Context, serviceName string) ([]ServiceInstancer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances, ok := r.services[serviceName]
	if !ok {
		return nil, ErrNotFound
	}

	result := make([]ServiceInstancer, len(instances))
	copy(result, instances)

	return result, nil
}

// Close closes the registry
func (r *DefaultRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.services = make(map[string][]ServiceInstancer)
	return nil
}
