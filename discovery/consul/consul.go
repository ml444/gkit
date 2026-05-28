package consul

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/ml444/gkit/discovery"
)

var _ discovery.ServiceRegistry = (*ConsulRegistry)(nil)

type agentAPI interface {
	ServiceRegister(*consulapi.AgentServiceRegistration) error
	ServiceDeregister(string) error
	UpdateTTL(string, string, string) error
}

type healthAPI interface {
	Service(string, string, bool, *consulapi.QueryOptions) ([]*consulapi.ServiceEntry, *consulapi.QueryMeta, error)
}

// ConsulRegistry implements discovery.ServiceRegistry using Consul
type ConsulRegistry struct {
	agent        agentAPI
	health       healthAPI
	serviceMap   sync.Map
	deregisterCh chan string
	healthCheck  bool
	ttl          int
	ttlCancels   sync.Map
	closeOnce    sync.Once
	closed       chan struct{}
}

type ConsulRegistryOption func(*ConsulRegistry)

func WithHealthCheck(enable bool) ConsulRegistryOption {
	return func(r *ConsulRegistry) {
		r.healthCheck = enable
	}
}

func WithTTL(ttl int) ConsulRegistryOption {
	return func(r *ConsulRegistry) {
		r.ttl = ttl
	}
}

func NewConsulRegistry(config *consulapi.Config, options ...ConsulRegistryOption) (*ConsulRegistry, error) {
	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consul client: %w", err)
	}
	return NewConsulRegistryWithAPI(client.Agent(), client.Health(), options...)
}

func NewConsulRegistryWithAPI(agent agentAPI, health healthAPI, options ...ConsulRegistryOption) (*ConsulRegistry, error) {
	registry := &ConsulRegistry{
		agent:        agent,
		health:       health,
		deregisterCh: make(chan string, 100),
		healthCheck:  true,
		ttl:          60,
		closed:       make(chan struct{}),
	}

	for _, option := range options {
		option(registry)
	}

	registry.startMonitoring()
	return registry, nil
}

func copyInstances(instances []discovery.ServiceInstancer) []discovery.ServiceInstancer {
	result := make([]discovery.ServiceInstancer, len(instances))
	copy(result, instances)
	return result
}

func (r *ConsulRegistry) Register(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	service := &consulapi.AgentServiceRegistration{
		ID:      instance.GetID(),
		Name:    instance.GetName(),
		Address: instance.GetAddress(),
		Port:    instance.GetPort(),
	}

	if instance.GetMetadata() != nil {
		metaData, err := json.Marshal(instance.GetMetadata())
		if err == nil {
			service.Meta = map[string]string{
				"metadata": string(metaData),
			}
		}
	}

	if r.healthCheck {
		service.Check = &consulapi.AgentServiceCheck{
			TTL:                            fmt.Sprintf("%ds", r.ttl),
			DeregisterCriticalServiceAfter: fmt.Sprintf("%ds", r.ttl*2),
		}
	}

	if err := r.agent.ServiceRegister(service); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	if r.healthCheck {
		r.startTTLPings(instance.GetID())
	}

	r.serviceMap.Delete(instance.GetName())
	return nil
}

func (r *ConsulRegistry) Deregister(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	if cancel, ok := r.ttlCancels.LoadAndDelete(instance.GetID()); ok {
		cancel.(context.CancelFunc)()
	}

	if err := r.agent.ServiceDeregister(instance.GetID()); err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}

	r.serviceMap.Delete(instance.GetName())

	select {
	case r.deregisterCh <- instance.GetName():
	case <-r.closed:
	}

	return nil
}

func (r *ConsulRegistry) GetServiceInstances(ctx context.Context, serviceName string) ([]discovery.ServiceInstancer, error) {
	if instances, ok := r.serviceMap.Load(serviceName); ok {
		return copyInstances(instances.([]discovery.ServiceInstancer)), nil
	}

	services, _, err := r.health.Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service instances: %w", err)
	}

	if len(services) == 0 {
		return nil, discovery.ErrNotFound
	}

	instances := make([]discovery.ServiceInstancer, 0, len(services))
	for _, service := range services {
		var metadata map[string]string
		if metaStr, ok := service.Service.Meta["metadata"]; ok {
			_ = json.Unmarshal([]byte(metaStr), &metadata)
		}

		instances = append(instances, &discovery.ServiceInstance{
			ID:       service.Service.ID,
			Name:     service.Service.Service,
			Address:  service.Service.Address,
			Port:     service.Service.Port,
			Metadata: metadata,
		})
	}

	r.serviceMap.Store(serviceName, instances)
	return copyInstances(instances), nil
}

func (r *ConsulRegistry) Close() error {
	r.closeOnce.Do(func() {
		close(r.closed)
		close(r.deregisterCh)
	})
	r.ttlCancels.Range(func(key, value any) bool {
		value.(context.CancelFunc)()
		r.ttlCancels.Delete(key)
		return true
	})
	return nil
}

func (r *ConsulRegistry) startMonitoring() {
	go func() {
		for serviceName := range r.deregisterCh {
			r.serviceMap.Delete(serviceName)
		}
	}()
}

func (r *ConsulRegistry) startTTLPings(serviceID string) {
	if cancel, ok := r.ttlCancels.Load(serviceID); ok {
		cancel.(context.CancelFunc)()
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.ttlCancels.Store(serviceID, cancel)

	go func() {
		ticker := time.NewTicker(time.Duration(r.ttl/2) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-r.closed:
				return
			case <-ticker.C:
				_ = r.agent.UpdateTTL(fmt.Sprintf("service:%s", serviceID), "", "passing")
			}
		}
	}()
}
