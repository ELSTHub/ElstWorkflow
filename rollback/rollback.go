// Package rollback 提供了 Saga 模式的回滚管理实现。
// 支持回滚栈、部分补偿和跳过回滚等功能。
package rollback

import (
	"fmt"
	"sync"

	"github.com/ELSTHub/elstworkflow/core"
)

// RollbackError 表示回滚错误
type RollbackError struct {
	// NodeName 回滚失败的节点名称
	NodeName string
	// Err 原始错误
	Err error
	// Partial 是否部分回滚成功
	Partial bool
}

// Error 返回错误信息
func (e *RollbackError) Error() string {
	if e.Partial {
		return fmt.Sprintf("partial rollback failed for node %s: %v", e.NodeName, e.Err)
	}
	return fmt.Sprintf("rollback failed for node %s: %v", e.NodeName, e.Err)
}

// Unwrap 返回原始错误
func (e *RollbackError) Unwrap() error {
	return e.Err
}

// Compensation 表示一个补偿操作
type Compensation struct {
	// NodeName 节点名称
	NodeName string
	// RollbackFunc 回滚函数
	RollbackFunc core.RollbackFunc
	// SkipRollback 是否跳过回滚
	SkipRollback bool
	// RollbackOnly 是否只执行回滚（不执行节点）
	RollbackOnly bool
	// Context 节点执行时的上下文
	Context core.Context
}

// RollbackManager 定义回滚管理器接口
type RollbackManager interface {
	// Push 添加补偿操作到回滚栈
	Push(compensation Compensation)
	// Pop 从回滚栈弹出补偿操作
	Pop() (Compensation, bool)
	// Rollback 执行所有补偿操作
	Rollback(ctx core.Context) error
	// RollbackUntil 执行到指定节点为止的补偿操作
	RollbackUntil(nodeName string, ctx core.Context) error
	// RollbackOnly 执行指定节点的回滚
	RollbackOnly(nodeName string, ctx core.Context) error
	// SkipRollback 跳过指定节点的回滚
	SkipRollback(nodeName string)
	// Clear 清空回滚栈
	Clear()
	// Size 返回回滚栈大小
	Size() int
	// HasCompensation 检查是否有指定节点的补偿操作
	HasCompensation(nodeName string) bool
}

// rollbackManager 是 RollbackManager 接口的默认实现
type rollbackManager struct {
	mu            sync.Mutex
	compensations []Compensation
	skipNodes     map[string]bool
}

// NewRollbackManager 创建新的回滚管理器
func NewRollbackManager() RollbackManager {
	return &rollbackManager{
		compensations: make([]Compensation, 0),
		skipNodes:     make(map[string]bool),
	}
}

// Push 添加补偿操作到回滚栈
func (m *rollbackManager) Push(compensation Compensation) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.compensations = append(m.compensations, compensation)
}

// Pop 从回滚栈弹出补偿操作
func (m *rollbackManager) Pop() (Compensation, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.compensations) == 0 {
		return Compensation{}, false
	}

	compensation := m.compensations[len(m.compensations)-1]
	m.compensations = m.compensations[:len(m.compensations)-1]
	return compensation, true
}

// Rollback 执行所有补偿操作
func (m *rollbackManager) Rollback(ctx core.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	successCount := 0

	// 从后往前执行回滚
	for i := len(m.compensations) - 1; i >= 0; i-- {
		compensation := m.compensations[i]

		// 检查是否跳过回滚
		if compensation.SkipRollback || m.skipNodes[compensation.NodeName] {
			continue
		}

		// 执行回滚函数
		if compensation.RollbackFunc != nil {
			if err := compensation.RollbackFunc(ctx); err != nil {
				lastErr = &RollbackError{
					NodeName: compensation.NodeName,
					Err:      err,
					Partial:  successCount > 0,
				}
			} else {
				successCount++
			}
		}
	}

	return lastErr
}

// RollbackUntil 执行到指定节点为止的补偿操作
// 从栈顶开始回滚，直到（并包括）指定节点为止
func (m *rollbackManager) RollbackUntil(nodeName string, ctx core.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	successCount := 0
	found := false

	// 从后往前执行回滚
	for i := len(m.compensations) - 1; i >= 0; i-- {
		compensation := m.compensations[i]

		// 检查是否跳过回滚
		if compensation.SkipRollback || m.skipNodes[compensation.NodeName] {
			// 如果是目标节点，停止
			if compensation.NodeName == nodeName {
				found = true
				break
			}
			continue
		}

		// 执行回滚函数
		if compensation.RollbackFunc != nil {
			if err := compensation.RollbackFunc(ctx); err != nil {
				lastErr = &RollbackError{
					NodeName: compensation.NodeName,
					Err:      err,
					Partial:  successCount > 0,
				}
			} else {
				successCount++
			}
		}

		// 如果是目标节点，停止
		if compensation.NodeName == nodeName {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("node %s not found in rollback stack", nodeName)
	}

	return lastErr
}

// RollbackOnly 执行指定节点的回滚
func (m *rollbackManager) RollbackOnly(nodeName string, ctx core.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 查找指定节点的补偿操作
	for _, compensation := range m.compensations {
		if compensation.NodeName == nodeName {
			if compensation.RollbackFunc != nil {
				return compensation.RollbackFunc(ctx)
			}
			return nil
		}
	}

	return fmt.Errorf("node %s not found in rollback stack", nodeName)
}

// SkipRollback 跳过指定节点的回滚
func (m *rollbackManager) SkipRollback(nodeName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.skipNodes[nodeName] = true
}

// Clear 清空回滚栈
func (m *rollbackManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.compensations = make([]Compensation, 0)
	m.skipNodes = make(map[string]bool)
}

// Size 返回回滚栈大小
func (m *rollbackManager) Size() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.compensations)
}

// HasCompensation 检查是否有指定节点的补偿操作
func (m *rollbackManager) HasCompensation(nodeName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, compensation := range m.compensations {
		if compensation.NodeName == nodeName {
			return true
		}
	}
	return false
}

// ExecuteWithRollback 执行操作，失败时自动回滚
func ExecuteWithRollback(manager RollbackManager, ctx core.Context, fn func() error, rollbackFn core.RollbackFunc, nodeName string) error {
	// 添加补偿操作
	manager.Push(Compensation{
		NodeName:     nodeName,
		RollbackFunc: rollbackFn,
		Context:      ctx,
	})

	// 执行操作
	if err := fn(); err != nil {
		// 执行回滚
		if rollbackErr := manager.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("operation failed: %v, rollback also failed: %v", err, rollbackErr)
		}
		return err
	}

	return nil
}

// ExecuteWithPartialRollback 执行操作，失败时执行部分回滚
func ExecuteWithPartialRollback(manager RollbackManager, ctx core.Context, fn func() error, rollbackFn core.RollbackFunc, nodeName string) error {
	// 添加补偿操作
	manager.Push(Compensation{
		NodeName:     nodeName,
		RollbackFunc: rollbackFn,
		Context:      ctx,
	})

	// 执行操作
	if err := fn(); err != nil {
		// 执行部分回滚（不返回回滚错误）
		manager.Rollback(ctx)
		return err
	}

	return nil
}
