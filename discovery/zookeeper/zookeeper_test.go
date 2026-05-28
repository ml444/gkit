package zookeeper

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/go-zookeeper/zk"
	"github.com/ml444/gkit/discovery"
)

type fakeZKNode struct {
	data     []byte
	children map[string]*fakeZKNode
	ephemeral bool
}

type fakeZKConn struct {
	mu   sync.RWMutex
	root *fakeZKNode
}

func newFakeZKConn() *fakeZKConn {
	return &fakeZKConn{root: &fakeZKNode{children: make(map[string]*fakeZKNode)}}
}

func (f *fakeZKConn) node(path string) (*fakeZKNode, bool) {
	if path == "" || path == "/" {
		return f.root, true
	}
	parts := splitPath(path)
	cur := f.root
	for _, part := range parts {
		child, ok := cur.children[part]
		if !ok {
			return nil, false
		}
		cur = child
	}
	return cur, true
}

func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	return strings.Split(path, "/")
}

func (f *fakeZKConn) Exists(path string) (bool, *zk.Stat, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	_, ok := f.node(path)
	return ok, nil, nil
}

func (f *fakeZKConn) Create(path string, data []byte, flags int32, _ []zk.ACL) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	parts := splitPath(path)
	if len(parts) == 0 {
		return "", zk.ErrNodeExists
	}

	cur := f.root
	for i, part := range parts {
		child, ok := cur.children[part]
		if !ok {
			child = &fakeZKNode{children: make(map[string]*fakeZKNode), ephemeral: flags&zk.FlagEphemeral != 0}
			if i == len(parts)-1 {
				child.data = data
			}
			cur.children[part] = child
			cur = child
			continue
		}
		if i == len(parts)-1 {
			return "", zk.ErrNodeExists
		}
		cur = child
	}
	return path, nil
}

func (f *fakeZKConn) Delete(path string, _ int32) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	parts := splitPath(path)
	if len(parts) == 0 {
		return zk.ErrNoNode
	}

	cur := f.root
	for i, part := range parts {
		child, ok := cur.children[part]
		if !ok {
			return zk.ErrNoNode
		}
		if i == len(parts)-1 {
			delete(cur.children, part)
			return nil
		}
		cur = child
	}
	return zk.ErrNoNode
}

func (f *fakeZKConn) Get(path string) ([]byte, *zk.Stat, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	node, ok := f.node(path)
	if !ok {
		return nil, nil, zk.ErrNoNode
	}
	return node.data, nil, nil
}

func (f *fakeZKConn) Children(path string) ([]string, *zk.Stat, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	node, ok := f.node(path)
	if !ok {
		return nil, nil, zk.ErrNoNode
	}
	names := make([]string, 0, len(node.children))
	for name := range node.children {
		names = append(names, name)
	}
	return names, nil, nil
}

func (f *fakeZKConn) Close() {}

func TestZookeeperRegistry_RegisterGetDeregister(t *testing.T) {
	conn := newFakeZKConn()
	reg, err := NewZookeeperRegistryWithConn(conn)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ctx := context.Background()
	ins := &discovery.ServiceInstance{
		ID: "1", Name: "svc", Address: "127.0.0.1", Port: 8080,
	}
	if err := reg.Register(ctx, ins); err != nil {
		t.Fatalf("register: %v", err)
	}

	path := "/gkit/services/svc/1"
	if ok, _, _ := conn.Exists(path); !ok {
		t.Fatalf("expected instance node at %s", path)
	}

	instances, err := reg.GetServiceInstances(ctx, "svc")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(instances) != 1 || instances[0].GetAddress() != "127.0.0.1" {
		t.Fatalf("unexpected instances: %+v", instances)
	}

	if err := reg.Deregister(ctx, ins); err != nil {
		t.Fatalf("deregister: %v", err)
	}
	if ok, _, _ := conn.Exists(path); ok {
		t.Fatalf("expected node deleted at %s", path)
	}
	_, err = reg.GetServiceInstances(ctx, "svc")
	if !errors.Is(err, discovery.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestZookeeperRegistry_RegisterDeregisterSamePath(t *testing.T) {
	conn := newFakeZKConn()
	reg, err := NewZookeeperRegistryWithConn(conn)
	if err != nil {
		t.Fatalf("new registry: %v", err)
	}
	defer reg.Close()

	ins := &discovery.ServiceInstance{ID: "abc", Name: "demo", Address: "1.1.1.1", Port: 80}
	if err := reg.Register(context.Background(), ins); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := reg.Deregister(context.Background(), ins); err != nil {
		t.Fatalf("deregister should use same path as register: %v", err)
	}
}
