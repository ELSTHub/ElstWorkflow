package registry

import (
	"testing"

	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/node"
)

func TestNew(t *testing.T) {
	r := New()
	if r == nil {
		t.Fatal("New() returned nil")
	}
}

func TestRegister(t *testing.T) {
	r := New()

	factory := func(name string, opts ...node.Option) node.Node {
		return node.NewFuncNode(name, func(ctx core.Context) (interface{}, error) {
			return nil, nil
		})
	}

	if err := r.Register("test", factory); err != nil {
		t.Errorf("Register failed: %v", err)
	}
}

func TestRegisterEmptyName(t *testing.T) {
	r := New()

	factory := func(name string, opts ...node.Option) node.Node {
		return nil
	}

	if err := r.Register("", factory); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRegisterNilFactory(t *testing.T) {
	r := New()

	if err := r.Register("test", nil); err == nil {
		t.Error("expected error for nil factory")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	r := New()

	factory := func(name string, opts ...node.Option) node.Node {
		return nil
	}

	r.Register("test", factory)
	if err := r.Register("test", factory); err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestFind(t *testing.T) {
	r := New()

	factory := func(name string, opts ...node.Option) node.Node {
		return node.NewFuncNode(name, func(ctx core.Context) (interface{}, error) {
			return nil, nil
		})
	}

	r.Register("test", factory)

	found, ok := r.Find("test")
	if !ok {
		t.Error("expected to find factory")
	}
	if found == nil {
		t.Error("expected non-nil factory")
	}
}

func TestFindNotFound(t *testing.T) {
	r := New()

	_, ok := r.Find("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestCreate(t *testing.T) {
	r := New()

	factory := func(name string, opts ...node.Option) node.Node {
		return node.NewFuncNode(name, func(ctx core.Context) (interface{}, error) {
			return "created", nil
		})
	}

	r.Register("test", factory)

	n, err := r.Create("test")
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	if n == nil {
		t.Error("expected non-nil node")
	}
	if n.Name() != "test" {
		t.Errorf("expected name 'test', got '%s'", n.Name())
	}
}

func TestCreateNotFound(t *testing.T) {
	r := New()

	_, err := r.Create("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestList(t *testing.T) {
	r := New()

	factory := func(name string, opts ...node.Option) node.Node {
		return nil
	}

	r.Register("node1", factory)
	r.Register("node2", factory)
	r.Register("node3", factory)

	names := r.List()
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d", len(names))
	}
}

func TestHas(t *testing.T) {
	r := New()

	factory := func(name string, opts ...node.Option) node.Node {
		return nil
	}

	r.Register("test", factory)

	if !r.Has("test") {
		t.Error("expected Has to return true")
	}
	if r.Has("nonexistent") {
		t.Error("expected Has to return false")
	}
}

func TestUnregister(t *testing.T) {
	r := New()

	factory := func(name string, opts ...node.Option) node.Node {
		return nil
	}

	r.Register("test", factory)

	if err := r.Unregister("test"); err != nil {
		t.Errorf("Unregister failed: %v", err)
	}

	if r.Has("test") {
		t.Error("expected node to be unregistered")
	}
}

func TestUnregisterNotFound(t *testing.T) {
	r := New()

	if err := r.Unregister("nonexistent"); err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestGlobalRegistry(t *testing.T) {
	// 清理全局注册表
	defaultRegistry = New()

	factory := func(name string, opts ...node.Option) node.Node {
		return node.NewFuncNode(name, func(ctx core.Context) (interface{}, error) {
			return nil, nil
		})
	}

	if err := Register("global-test", factory); err != nil {
		t.Errorf("Register failed: %v", err)
	}

	if !Has("global-test") {
		t.Error("expected Has to return true")
	}

	found, ok := Find("global-test")
	if !ok || found == nil {
		t.Error("expected to find factory")
	}

	n, err := Create("global-test")
	if err != nil || n == nil {
		t.Error("expected to create node")
	}

	names := List()
	if len(names) != 1 {
		t.Errorf("expected 1 name, got %d", len(names))
	}

	if err := Unregister("global-test"); err != nil {
		t.Errorf("Unregister failed: %v", err)
	}
}

func TestRegisterFuncNode(t *testing.T) {
	defaultRegistry = New()

	if err := RegisterFuncNode("func-node", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	}); err != nil {
		t.Errorf("RegisterFuncNode failed: %v", err)
	}

	if !Has("func-node") {
		t.Error("expected Has to return true")
	}
}

func TestRegistryInterface(t *testing.T) {
	// 确保 registry 实现了 Registry 接口
	var _ Registry = New()
}

// 基准测试
func BenchmarkRegister(b *testing.B) {
	r := New()
	factory := func(name string, opts ...node.Option) node.Node {
		return nil
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Register("test", factory)
	}
}

func BenchmarkFind(b *testing.B) {
	r := New()
	factory := func(name string, opts ...node.Option) node.Node {
		return nil
	}
	r.Register("test", factory)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Find("test")
	}
}

func BenchmarkCreate(b *testing.B) {
	r := New()
	factory := func(name string, opts ...node.Option) node.Node {
		return node.NewFuncNode(name, func(ctx core.Context) (interface{}, error) {
			return nil, nil
		})
	}
	r.Register("test", factory)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Create("test")
	}
}
