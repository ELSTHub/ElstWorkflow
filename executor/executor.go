// Package executor 提供了节点执行器的实现。
// 执行器负责执行节点，处理重试、超时和中间件。
package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/node"
	"github.com/ELSTHub/elstworkflow/retry"
	"github.com/ELSTHub/elstworkflow/rollback"
	"github.com/ELSTHub/elstworkflow/timeout"
)

// Executor 定义执行器接口
type Executor interface {
	// Execute 执行节点
	Execute(ctx core.Context, n node.Node) (*core.NodeResult, error)
	// ExecuteWithRollback 执行节点，失败时回滚
	ExecuteWithRollback(ctx core.Context, n node.Node, rollbackMgr rollback.RollbackManager) (*core.NodeResult, error)
}

// Middleware 定义执行器中间件类型
type Middleware func(next ExecuteFunc) ExecuteFunc

// ExecuteFunc 定义执行函数类型
type ExecuteFunc func(ctx core.Context, n node.Node) (*core.NodeResult, error)

// executor 是 Executor 接口的默认实现
type executor struct {
	mu          sync.RWMutex
	middlewares []Middleware
	retryer     retry.Retryer
	timeoutMgr  timeout.TimeoutManager
}

// New 创建新的执行器
func New(opts ...Option) Executor {
	e := &executor{
		middlewares: make([]Middleware, 0),
		timeoutMgr:  timeout.NewTimeoutManager(),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Option 定义执行器选项
type Option func(*executor)

// WithMiddleware 添加中间件
func WithMiddleware(middleware ...Middleware) Option {
	return func(e *executor) {
		e.middlewares = append(e.middlewares, middleware...)
	}
}

// WithRetryer 设置重试器
func WithRetryer(retryer retry.Retryer) Option {
	return func(e *executor) {
		e.retryer = retryer
	}
}

// Execute 执行节点
func (e *executor) Execute(ctx core.Context, n node.Node) (*core.NodeResult, error) {
	e.mu.RLock()
	middlewares := make([]Middleware, len(e.middlewares))
	copy(middlewares, e.middlewares)
	retryer := e.retryer
	timeoutMgr := e.timeoutMgr
	e.mu.RUnlock()

	// 构建执行链
	handler := e.buildChain(middlewares)

	// 执行节点
	return e.executeWithRetryAndTimeout(ctx, n, handler, retryer, timeoutMgr)
}

// ExecuteWithRollback 执行节点，失败时回滚
func (e *executor) ExecuteWithRollback(ctx core.Context, n node.Node, rollbackMgr rollback.RollbackManager) (*core.NodeResult, error) {
	// 注册回滚函数
	opts := n.Options()
	if opts != nil {
		rollbackMgr.Push(rollback.Compensation{
			NodeName: n.Name(),
			RollbackFunc: func(ctx core.Context) error {
				return n.Rollback(ctx)
			},
			Context: ctx,
		})
	}

	// 执行节点
	result, err := e.Execute(ctx, n)
	if err != nil {
		// 执行回滚
		if rollbackErr := rollbackMgr.Rollback(ctx); rollbackErr != nil {
			return nil, fmt.Errorf("execution failed: %v, rollback also failed: %v", err, rollbackErr)
		}
		return result, err
	}

	return result, nil
}

// buildChain 构建执行链
func (e *executor) buildChain(middlewares []Middleware) ExecuteFunc {
	// 最终的执行函数
	handler := func(ctx core.Context, n node.Node) (*core.NodeResult, error) {
		return e.executeNode(ctx, n)
	}

	// 从后往前包装中间件
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}

// executeNode 执行单个节点
func (e *executor) executeNode(ctx core.Context, n node.Node) (*core.NodeResult, error) {
	startTime := time.Now()

	// 验证节点
	if err := n.Validate(); err != nil {
		return &core.NodeResult{
			Status:    core.NodeFailed,
			Error:     err,
			StartTime: startTime,
			EndTime:   time.Now(),
			Duration:  time.Since(startTime),
		}, err
	}

	// 执行节点
	output, err := n.Execute(ctx)
	endTime := time.Now()

	if err != nil {
		return &core.NodeResult{
			Status:    core.NodeFailed,
			Error:     err,
			StartTime: startTime,
			EndTime:   endTime,
			Duration:  endTime.Sub(startTime),
		}, err
	}

	return &core.NodeResult{
		Status:    core.NodeCompleted,
		Output:    output,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}

// executeWithRetryAndTimeout 执行带重试和超时的节点
func (e *executor) executeWithRetryAndTimeout(ctx core.Context, n node.Node, handler ExecuteFunc, retryer retry.Retryer, timeoutMgr timeout.TimeoutManager) (*core.NodeResult, error) {
	opts := n.Options()

	// 确定超时时间
	var timeoutDuration time.Duration
	if opts != nil && opts.TimeoutPolicy != nil {
		timeoutDuration = opts.TimeoutPolicy.Timeout
	}

	// 确定重试策略
	var nodeRetryer retry.Retryer
	if retryer != nil {
		nodeRetryer = retryer
	} else if opts != nil && opts.RetryPolicy != nil {
		nodeRetryer = retry.NewRetryer(opts.RetryPolicy)
	}

	// 执行函数
	executeFn := func() (interface{}, error) {
		if timeoutDuration > 0 {
			return timeoutMgr.WithTimeout(timeoutDuration, func(tctx context.Context) (interface{}, error) {
				return handler(ctx, n)
			})
		}
		return handler(ctx, n)
	}

	// 带重试执行
	if nodeRetryer != nil {
		result, err := retry.Execute(nodeRetryer, executeFn)
		if err != nil {
			return nil, err
		}
		if nr, ok := result.(*core.NodeResult); ok {
			return nr, nil
		}
		return &core.NodeResult{
			Status: core.NodeCompleted,
			Output: result,
		}, nil
	}

	// 无重试执行
	result, err := executeFn()
	if err != nil {
		return nil, err
	}
	if nr, ok := result.(*core.NodeResult); ok {
		return nr, nil
	}
	return &core.NodeResult{
		Status: core.NodeCompleted,
		Output: result,
	}, nil
}
