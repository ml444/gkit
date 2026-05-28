package k8s

import (
	"context"
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ml444/gkit/discovery"
)

func newTestRegistry() *K8sRegistry {
	return &K8sRegistry{
		services: make(map[string][]discovery.ServiceInstancer),
	}
}

func TestK8sRegistry_HandleServiceUpdate(t *testing.T) {
	reg := newTestRegistry()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.96.0.1",
			Ports:     []corev1.ServicePort{{Port: 8080}},
		},
	}

	reg.handleServiceUpdate(svc)
	instances, err := reg.GetServiceInstances(context.Background(), "api")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 || instances[0].GetAddress() != "10.96.0.1" {
		t.Fatalf("unexpected instances: %+v", instances)
	}
}

func TestK8sRegistry_HandleServiceDelete(t *testing.T) {
	reg := newTestRegistry()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "api"},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.96.0.1",
			Ports:     []corev1.ServicePort{{Port: 8080}},
		},
	}
	reg.handleServiceUpdate(svc)
	reg.handleServiceDelete(svc)

	_, err := reg.GetServiceInstances(context.Background(), "api")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestK8sRegistry_HandlePodUpdateUsesContainerPort(t *testing.T) {
	reg := newTestRegistry()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "pod-1",
			Labels: map[string]string{"app": "worker"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.0.0.5",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Ports: []corev1.ContainerPort{{ContainerPort: 9090}},
			}},
		},
	}

	reg.handlePodUpdate(pod)
	instances, err := reg.GetServiceInstances(context.Background(), "worker")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 || instances[0].GetPort() != 9090 {
		t.Fatalf("unexpected instances: %+v", instances)
	}
}

func TestK8sRegistry_HandlePodDeleteTombstone(t *testing.T) {
	reg := newTestRegistry()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "pod-1",
			Labels: map[string]string{"app": "worker"},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.5"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Ports: []corev1.ContainerPort{{ContainerPort: 8080}}}},
		},
	}
	reg.handlePodUpdate(pod)
	reg.handlePodDelete(pod)

	_, err := reg.GetServiceInstances(context.Background(), "worker")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestK8sRegistry_RegisterDeregisterLocalCache(t *testing.T) {
	reg := newTestRegistry()
	ctx := context.Background()
	ins := &discovery.ServiceInstance{ID: "1", Name: "local", Address: "127.0.0.1", Port: 8080}

	if err := reg.Register(ctx, ins); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := reg.Deregister(ctx, ins); err != nil {
		t.Fatalf("deregister: %v", err)
	}
	_, err := reg.GetServiceInstances(ctx, "local")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
