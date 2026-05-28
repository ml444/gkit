package nacos

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"

	"github.com/ml444/gkit/discovery"
)

var _ discovery.ServiceRegistry = (*NacosRegistry)(nil)

type namingClient interface {
	RegisterInstance(param vo.RegisterInstanceParam) (bool, error)
	DeregisterInstance(param vo.DeregisterInstanceParam) (bool, error)
	SelectInstances(param vo.SelectInstancesParam) ([]model.Instance, error)
	Subscribe(param *vo.SubscribeParam) error
}

type NacosRegistry struct {
	client       namingClient
	serviceMap   sync.Map
	deregisterCh chan string
	timeout      time.Duration
	closeOnce    sync.Once
}

type NacosRegistryOption func(*NacosRegistry)

func WithTimeout(timeout time.Duration) NacosRegistryOption {
	return func(r *NacosRegistry) {
		r.timeout = timeout
	}
}

func NewNacosRegistry(serverConfigs []constant.ServerConfig, clientConfig constant.ClientConfig, options ...NacosRegistryOption) (*NacosRegistry, error) {
	client, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create nacos client: %w", err)
	}
	return NewNacosRegistryWithClient(client, options...)
}

func NewNacosRegistryWithClient(client naming_client.INamingClient, options ...NacosRegistryOption) (*NacosRegistry, error) {
	return newNacosRegistry(client, options...)
}

func newNacosRegistry(client namingClient, options ...NacosRegistryOption) (*NacosRegistry, error) {
	registry := &NacosRegistry{
		client:       client,
		deregisterCh: make(chan string, 100),
		timeout:      5 * time.Second,
	}

	for _, option := range options {
		option(registry)
	}

	go func() {
		for serviceName := range registry.deregisterCh {
			registry.serviceMap.Delete(serviceName)
		}
	}()

	return registry, nil
}

func copyInstances(instances []discovery.ServiceInstancer) []discovery.ServiceInstancer {
	result := make([]discovery.ServiceInstancer, len(instances))
	copy(result, instances)
	return result
}

func toDiscoveryInstances(serviceName string, instances []model.Instance) []discovery.ServiceInstancer {
	discoveryInstances := make([]discovery.ServiceInstancer, 0, len(instances))
	for _, instance := range instances {
		discoveryInstances = append(discoveryInstances, &discovery.ServiceInstance{
			ID:       fmt.Sprintf("%s:%d", instance.Ip, instance.Port),
			Name:     serviceName,
			Address:  instance.Ip,
			Port:     int(instance.Port),
			Metadata: instance.Metadata,
		})
	}
	return discoveryInstances
}

func (r *NacosRegistry) Register(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	nacosMetadata := make(map[string]string)
	if instance.GetMetadata() != nil {
		for k, v := range instance.GetMetadata() {
			nacosMetadata[k] = v
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	success, err := r.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          instance.GetAddress(),
		Port:        uint64(instance.GetPort()),
		ServiceName: instance.GetName(),
		Weight:      1.0,
		Enable:      true,
		Healthy:     true,
		Metadata:    nacosMetadata,
		ClusterName: "DEFAULT",
	})
	if err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}
	if !success {
		return fmt.Errorf("failed to register service: unknown error")
	}

	r.serviceMap.Delete(instance.GetName())
	return nil
}

func (r *NacosRegistry) Deregister(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	success, err := r.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          instance.GetAddress(),
		Port:        uint64(instance.GetPort()),
		ServiceName: instance.GetName(),
		Cluster:     "DEFAULT",
	})
	if err != nil {
		return fmt.Errorf("failed to deregister service: %w", err)
	}
	if !success {
		return fmt.Errorf("failed to deregister service: unknown error")
	}

	r.serviceMap.Delete(instance.GetName())

	select {
	case r.deregisterCh <- instance.GetName():
	default:
	}
	return nil
}

func (r *NacosRegistry) GetServiceInstances(ctx context.Context, serviceName string) ([]discovery.ServiceInstancer, error) {
	if instances, ok := r.serviceMap.Load(serviceName); ok {
		return copyInstances(instances.([]discovery.ServiceInstancer)), nil
	}

	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	instances, err := r.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get service instances: %w", err)
	}
	if len(instances) == 0 {
		return nil, discovery.ErrNotFound
	}

	discoveryInstances := toDiscoveryInstances(serviceName, instances)
	r.serviceMap.Store(serviceName, discoveryInstances)
	return copyInstances(discoveryInstances), nil
}

func (r *NacosRegistry) Close() error {
	r.closeOnce.Do(func() {
		close(r.deregisterCh)
	})
	return nil
}

func (r *NacosRegistry) Subscribe(serviceName string, listener func([]discovery.ServiceInstancer)) error {
	return r.client.Subscribe(&vo.SubscribeParam{
		ServiceName: serviceName,
		SubscribeCallback: func(services []model.Instance, err error) {
			if err != nil {
				return
			}
			discoveryInstances := toDiscoveryInstances(serviceName, services)
			r.serviceMap.Store(serviceName, discoveryInstances)
			if listener != nil {
				listener(copyInstances(discoveryInstances))
			}
		},
	})
}
