// Package context 提供线程安全的工作流上下文实现。
// 上下文用于在工作流节点之间传递数据，支持任意类型的键值存储。
package context

import (
	"fmt"
	"sync"

	"github.com/ELSTHub/elstworkflow/core"
)

// workflowContext 是 core.Context 接口的默认实现
type workflowContext struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// New 创建一个新的工作流上下文
func New() core.Context {
	return &workflowContext{
		data: make(map[string]interface{}),
	}
}

// Put 设置键值对
func (c *workflowContext) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Get 获取键对应的值
func (c *workflowContext) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

// GetTyped 获取键对应的值并尝试转换为指定类型
func (c *workflowContext) GetTyped(key string, target interface{}) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.data[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}

	// 如果target是nil，直接返回值
	if target == nil {
		return val, nil
	}

	return val, nil
}

// GetT 是泛型版本的Get方法，提供类型安全的值获取
func GetT[T any](ctx core.Context, key string) (T, error) {
	val, ok := ctx.Get(key)
	if !ok {
		var zero T
		return zero, fmt.Errorf("key %s not found", key)
	}

	typedVal, ok := val.(T)
	if !ok {
		var zero T
		return zero, fmt.Errorf("value for key %s is not of expected type", key)
	}

	return typedVal, nil
}

// MustGetT 是泛型版本的Get方法，如果键不存在或类型不匹配会panic
func MustGetT[T any](ctx core.Context, key string) T {
	val, err := GetT[T](ctx, key)
	if err != nil {
		panic(err)
	}
	return val
}

// Delete 删除键值对
func (c *workflowContext) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Keys 返回所有键
func (c *workflowContext) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

// Clone 克隆上下文
func (c *workflowContext) Clone() core.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cloned := &workflowContext{
		data: make(map[string]interface{}, len(c.data)),
	}

	for k, v := range c.data {
		cloned.data[k] = v
	}

	return cloned
}

// Has 检查键是否存在
func (c *workflowContext) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.data[key]
	return ok
}

// Len 返回键值对数量
func (c *workflowContext) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Clear 清空所有键值对
func (c *workflowContext) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]interface{})
}

// String 返回上下文的字符串表示
func (c *workflowContext) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("Context{keys: %d}", len(c.data))
}
