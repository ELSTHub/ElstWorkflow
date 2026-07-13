// Package timeout 提供了超时策略的实现。
// 支持工作流超时、节点超时和上下文取消。
package timeout

import (
	"context"
	"fmt"
	"time"

	"github.com/elstworkflow/core"
)

// TimeoutError 表示超时错误
type TimeoutError struct {
	// Operation 超时的操作名称
	Operation string
	// Timeout 超时时间
	Timeout time.Duration
	// Elapsed 已经过时间
	Elapsed time.Duration
}

// Error 返回错误信息
func (e *TimeoutError) Error() string {
	return fmt.Sprintf("operation %s timed out after %v (elapsed: %v)", e.Operation, e.Timeout, e.Elapsed)
}

// TimeoutManager 定义超时管理器接口
type TimeoutManager interface {
	// WithTimeout 执行带超时的操作
	WithTimeout(timeout time.Duration, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
	// WithTimeoutPolicy 使用超时策略执行操作
	WithTimeoutPolicy(policy *core.TimeoutPolicy, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
	// CreateContext 创建带超时的上下文
	CreateContext(timeout time.Duration) (context.Context, context.CancelFunc)
	// CreateContextWithParent 创建带父上下文和超时的上下文
	CreateContextWithParent(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc)
}

// timeoutManager 是 TimeoutManager 接口的默认实现
type timeoutManager struct{}

// NewTimeoutManager 创建新的超时管理器
func NewTimeoutManager() TimeoutManager {
	return &timeoutManager{}
}

// WithTimeout 执行带超时的操作
func (m *timeoutManager) WithTimeout(timeout time.Duration, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	if timeout <= 0 {
		return fn(context.Background())
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		value interface{}
		err   error
	}

	done := make(chan result, 1)
	go func() {
		value, err := fn(ctx)
		done <- result{value, err}
	}()

	select {
	case res := <-done:
		return res.value, res.err
	case <-ctx.Done():
		return nil, &TimeoutError{
			Operation: "operation",
			Timeout:   timeout,
			Elapsed:   timeout,
		}
	}
}

// WithTimeoutPolicy 使用超时策略执行操作
func (m *timeoutManager) WithTimeoutPolicy(policy *core.TimeoutPolicy, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	if policy == nil {
		return fn(context.Background())
	}

	return m.WithTimeout(policy.Timeout, fn)
}

// CreateContext 创建带超时的上下文
func (m *timeoutManager) CreateContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// CreateContextWithParent 创建带父上下文和超时的上下文
func (m *timeoutManager) CreateContextWithParent(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, timeout)
}

// ExecuteWithTimeout 执行带超时的操作（便捷函数）
func ExecuteWithTimeout(timeout time.Duration, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	manager := NewTimeoutManager()
	return manager.WithTimeout(timeout, fn)
}

// ExecuteWithTimeoutPolicy 使用超时策略执行操作（便捷函数）
func ExecuteWithTimeoutPolicy(policy *core.TimeoutPolicy, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	manager := NewTimeoutManager()
	return manager.WithTimeoutPolicy(policy, fn)
}

// IsTimeoutError 检查错误是否是超时错误
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*TimeoutError)
	return ok
}

// WrapWithTimeout 包装函数，添加超时功能
func WrapWithTimeout(fn func() (interface{}, error), timeout time.Duration) func() (interface{}, error) {
	return func() (interface{}, error) {
		return ExecuteWithTimeout(timeout, func(ctx context.Context) (interface{}, error) {
			return fn()
		})
	}
}

// WrapWithTimeoutPolicy 包装函数，使用超时策略
func WrapWithTimeoutPolicy(fn func() (interface{}, error), policy *core.TimeoutPolicy) func() (interface{}, error) {
	return func() (interface{}, error) {
		return ExecuteWithTimeoutPolicy(policy, func(ctx context.Context) (interface{}, error) {
			return fn()
		})
	}
}

// WrapWithContext 包装函数，添加上下文支持
func WrapWithContext(fn func(ctx context.Context) (interface{}, error), timeout time.Duration) func() (interface{}, error) {
	return func() (interface{}, error) {
		return ExecuteWithTimeout(timeout, fn)
	}
}
