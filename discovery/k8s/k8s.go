package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ml444/gkit/discovery"
)

var _ discovery.ServiceRegistry = (*K8sRegistry)(nil)

// K8sRegistry implements discovery.ServiceRegistry using Kubernetes API
// 基于Kubernetes API的服务注册发现实现
type K8sRegistry struct {
	clientset       *kubernetes.Clientset
	namespace       string
	informerFactory informers.SharedInformerFactory
	mu              sync.RWMutex
	services        map[string][]discovery.ServiceInstancer
	stopCh          chan struct{}
}

// K8sRegistryOption is option for K8sRegistry
// K8sRegistry配置选项
type K8sRegistryOption func(*K8sRegistry)

// WithNamespace sets the namespace for K8sRegistry
// 设置Kubernetes命名空间
func WithNamespace(namespace string) K8sRegistryOption {
	return func(r *K8sRegistry) {
		r.namespace = namespace
	}
}

// NewK8sRegistry creates a new K8sRegistry
// 创建一个新的基于Kubernetes的服务注册中心
func NewK8sRegistry(options ...K8sRegistryOption) (*K8sRegistry, error) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		config, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		).ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return NewK8sRegistryWithClientset(clientset, options...)
}

// NewK8sRegistryWithClientset creates a K8sRegistry with an existing clientset.
func NewK8sRegistryWithClientset(clientset kubernetes.Interface, options ...K8sRegistryOption) (*K8sRegistry, error) {
	cs, ok := clientset.(*kubernetes.Clientset)
	if !ok {
		return nil, fmt.Errorf("clientset must be *kubernetes.Clientset")
	}

	registry := &K8sRegistry{
		clientset: cs,
		namespace: "default",
		services:  make(map[string][]discovery.ServiceInstancer),
		stopCh:    make(chan struct{}),
	}

	for _, option := range options {
		option(registry)
	}

	registry.informerFactory = informers.NewSharedInformerFactoryWithOptions(cs, time.Minute*30, informers.WithNamespace(registry.namespace))
	registry.startInformers()

	return registry, nil
}

func (r *K8sRegistry) upsertInstance(serviceName string, instance discovery.ServiceInstancer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances := r.services[serviceName]
	for i, ins := range instances {
		if ins.GetID() == instance.GetID() {
			instances[i] = instance
			r.services[serviceName] = instances
			return
		}
	}
	r.services[serviceName] = append(instances, instance)
}

func (r *K8sRegistry) removeInstance(serviceName, instanceID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	instances, ok := r.services[serviceName]
	if !ok {
		return
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
}

func (r *K8sRegistry) deleteService(serviceName string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.services, serviceName)
}

func podPort(pod *corev1.Pod) int32 {
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			if port.ContainerPort > 0 {
				return port.ContainerPort
			}
		}
	}
	return 80
}

// startInformers starts the Kubernetes informers
// 启动Kubernetes informers以监听服务变化
func (r *K8sRegistry) startInformers() {
	// Service informer
	serviceInformer := r.informerFactory.Core().V1().Services().Informer()
	serviceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			svc := obj.(*corev1.Service)
			r.handleServiceUpdate(svc)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newSvc := newObj.(*corev1.Service)
			r.handleServiceUpdate(newSvc)
		},
		DeleteFunc: func(obj interface{}) {
			svc, ok := obj.(*corev1.Service)
			if !ok {
				// Handle tombstone events
				if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					if svc, ok := tombstone.Obj.(*corev1.Service); ok {
						r.handleServiceDelete(svc)
					}
				}
				return
			}
			r.handleServiceDelete(svc)
		},
	})

	// Pod informer
	podInformer := r.informerFactory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			r.handlePodUpdate(pod)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPod := newObj.(*corev1.Pod)
			r.handlePodUpdate(newPod)
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*corev1.Pod)
			if !ok {
				// Handle tombstone events
				if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					if pod, ok := tombstone.Obj.(*corev1.Pod); ok {
						r.handlePodDelete(pod)
					}
				}
				return
			}
			r.handlePodDelete(pod)
		},
	})

	// Start informers
	r.informerFactory.Start(r.stopCh)

	// Wait for informers to sync
	if !cache.WaitForCacheSync(r.stopCh, serviceInformer.HasSynced, podInformer.HasSynced) {
		// Log error but continue
		fmt.Println("Warning: not all informers synced")
	}
}

// handleServiceUpdate handles service update events
// 处理服务更新事件
func (r *K8sRegistry) handleServiceUpdate(svc *corev1.Service) {
	if len(svc.Spec.Ports) == 0 {
		return
	}
	if svc.Spec.ClusterIP != "None" && svc.Spec.ClusterIP != "" {
		instance := &discovery.ServiceInstance{
			ID:      svc.Name,
			Name:    svc.Name,
			Address: svc.Spec.ClusterIP,
			Port:    int(svc.Spec.Ports[0].Port),
			Metadata: map[string]string{
				"namespace": svc.Namespace,
				"labels":    labels.FormatLabels(svc.Labels),
			},
		}
		r.upsertInstance(svc.Name, instance)
	}
}

// handleServiceDelete handles service delete events
// 处理服务删除事件
func (r *K8sRegistry) handleServiceDelete(svc *corev1.Service) {
	r.deleteService(svc.Name)
}

// handlePodUpdate handles pod update events
// 处理Pod更新事件
func (r *K8sRegistry) handlePodUpdate(pod *corev1.Pod) {
	if pod.Status.Phase != corev1.PodRunning {
		return
	}

	for key, value := range pod.Labels {
		if strings.HasSuffix(key, "/service-name") || key == "app" {
			serviceName := value
			instance := &discovery.ServiceInstance{
				ID:      pod.Name,
				Name:    serviceName,
				Address: pod.Status.PodIP,
				Port:    int(podPort(pod)),
				Metadata: map[string]string{
					"namespace": pod.Namespace,
					"podName":   pod.Name,
				},
			}
			r.upsertInstance(serviceName, instance)
			break
		}
	}
}

// handlePodDelete handles pod delete events
// 处理Pod删除事件
func (r *K8sRegistry) handlePodDelete(pod *corev1.Pod) {
	for key, value := range pod.Labels {
		if strings.HasSuffix(key, "/service-name") || key == "app" {
			r.removeInstance(value, pod.Name)
			break
		}
	}
}

// Register registers a service instance
// 在Kubernetes环境中，服务注册通常由控制器自动处理，这里我们提供一个兼容接口
func (r *K8sRegistry) Register(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}
	r.upsertInstance(instance.GetName(), instance)
	return nil
}

func (r *K8sRegistry) Deregister(ctx context.Context, instance discovery.ServiceInstancer) error {
	if instance == nil {
		return fmt.Errorf("instance cannot be nil")
	}
	r.removeInstance(instance.GetName(), instance.GetID())
	return nil
}

func (r *K8sRegistry) GetServiceInstances(ctx context.Context, serviceName string) ([]discovery.ServiceInstancer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	instances, ok := r.services[serviceName]
	if !ok {
		return nil, discovery.ErrNotFound
	}

	result := make([]discovery.ServiceInstancer, len(instances))
	copy(result, instances)
	return result, nil
}

func (r *K8sRegistry) Close() error {
	close(r.stopCh)
	return nil
}