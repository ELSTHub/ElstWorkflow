// Package registry 提供了节点注册表的实现。
// 支持注册、查找和创建节点实例。
package registry

import (
	"fmt"
	"sync"

	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/node"
)

// Factory 定义节点工厂函数类型
type Factory func(name string, opts ...node.Option) node.Node

// Registry 定义注册表接口
type Registry interface {
	// Register 注册节点工厂
	Register(name string, factory Factory) error
	// Find 查找节点工厂
	Find(name string) (Factory, bool)
	// Create 创建节点实例
	Create(name string, opts ...node.Option) (node.Node, error)
	// List 列出所有已注册的节点名称
	List() []string
	// Has 检查是否已注册指定名称的节点
	Has(name string) bool
	// Unregister 注销节点工厂
	Unregister(name string) error
}

// registry 是 Registry 接口的默认实现
type registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
}

// New 创建新的注册表
func New() Registry {
	return &registry{
		factories: make(map[string]Factory),
	}
}

// Register 注册节点工厂
func (r *registry) Register(name string, factory Factory) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if name == "" {
		return fmt.Errorf("node name cannot be empty")
	}

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	if _, exists := r.factories[name]; exists {
		return fmt.Errorf("node %s already registered", name)
	}

	r.factories[name] = factory
	return nil
}

// Find 查找节点工厂
func (r *registry) Find(name string) (Factory, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[name]
	return factory, ok
}

// Create 创建节点实例
func (r *registry) Create(name string, opts ...node.Option) (node.Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	factory, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("node %s not registered", name)
	}

	return factory(name, opts...), nil
}

// List 列出所有已注册的节点名称
func (r *registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	return names
}

// Has 检查是否已注册指定名称的节点
func (r *registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.factories[name]
	return ok
}

// Unregister 注销节点工厂
func (r *registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.factories[name]; !exists {
		return fmt.Errorf("node %s not registered", name)
	}

	delete(r.factories, name)
	return nil
}

// 全局注册表
var defaultRegistry = New()

// Register 使用全局注册表注册节点工厂
func Register(name string, factory Factory) error {
	return defaultRegistry.Register(name, factory)
}

// Find 使用全局注册表查找节点工厂
func Find(name string) (Factory, bool) {
	return defaultRegistry.Find(name)
}

// Create 使用全局注册表创建节点实例
func Create(name string, opts ...node.Option) (node.Node, error) {
	return defaultRegistry.Create(name, opts...)
}

// List 使用全局注册表列出所有已注册的节点名称
func List() []string {
	return defaultRegistry.List()
}

// Has 使用全局注册表检查是否已注册指定名称的节点
func Has(name string) bool {
	return defaultRegistry.Has(name)
}

// Unregister 使用全局注册表注销节点工厂
func Unregister(name string) error {
	return defaultRegistry.Unregister(name)
}

// RegisterFuncNode 注册基于函数的节点
func RegisterFuncNode(name string, executeFn core.NodeFunc) error {
	return Register(name, func(n string, opts ...node.Option) node.Node {
		return node.NewFuncNode(n, executeFn, opts...)
	})
}

// RegisterFuncNodeWithRollback 注册带回滚函数的节点
func RegisterFuncNodeWithRollback(name string, executeFn core.NodeFunc, rollbackFn core.RollbackFunc) error {
	return Register(name, func(n string, opts ...node.Option) node.Node {
		return node.NewFuncNodeWithRollback(n, executeFn, rollbackFn, opts...)
	})
}
