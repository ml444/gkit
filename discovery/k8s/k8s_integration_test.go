//go:build integration

package k8s

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/ml444/gkit/discovery"
)

func TestK8sRegistry_IntegrationWithFakeClientset(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	reg, err := NewK8sRegistryWithClientset(clientset)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"},
		Spec: corev1.ServiceSpec{
			ClusterIP: "10.96.0.10",
			Ports:     []corev1.ServicePort{{Port: 8080}},
		},
	}
	if _, err := clientset.CoreV1().Services("default").Create(ctx, svc, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create service: %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		instances, err := reg.GetServiceInstances(ctx, "api")
		if err == nil && len(instances) == 1 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for informer sync, last err=%v", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "default",
			Labels:    map[string]string{"app": "worker"},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning, PodIP: "10.0.0.8"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Ports: []corev1.ContainerPort{{ContainerPort: 9000}}}},
		},
	}
	if _, err := clientset.CoreV1().Pods("default").Create(ctx, pod, metav1.CreateOptions{}); err != nil {
		t.Fatalf("create pod: %v", err)
	}

	deadline = time.Now().Add(5 * time.Second)
	for {
		instances, err := reg.GetServiceInstances(ctx, "worker")
		if err == nil && len(instances) == 1 && instances[0].GetPort() == 9000 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for pod informer, last err=%v instances=%v", err, instances)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func TestK8sRegistry_IntegrationRegisterLocal(t *testing.T) {
	reg := newTestRegistry()
	ins := &discovery.ServiceInstance{ID: "1", Name: "local", Address: "127.0.0.1", Port: 8080}
	if err := reg.Register(context.Background(), ins); err != nil {
		t.Fatalf("register: %v", err)
	}
	instances, err := reg.GetServiceInstances(context.Background(), "local")
	if err != nil || len(instances) != 1 {
		t.Fatalf("get: err=%v instances=%v", err, instances)
	}
}
