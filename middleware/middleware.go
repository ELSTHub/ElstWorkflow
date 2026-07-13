// Package middleware 提供了工作流中间件的实现。
// 支持日志、指标、追踪、审计和恢复等中间件。
package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/elstworkflow/core"
	"github.com/elstworkflow/node"
)

// Context 定义中间件上下文
type Context struct {
	// WorkflowName 工作流名称
	WorkflowName string
	// NodeName 节点名称
	NodeName string
	// StartTime 开始时间
	StartTime time.Time
	// Data 自定义数据
	Data map[string]interface{}
}

// Middleware 定义中间件接口
type Middleware interface {
	// Before 在节点执行前调用
	Before(ctx *Context, nodeCtx core.Context, n node.Node) error
	// After 在节点执行后调用
	After(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error)
	// OnError 在节点执行出错时调用
	OnError(ctx *Context, nodeCtx core.Context, n node.Node, err error)
}

// Chain 定义中间件链
type Chain struct {
	middlewares []Middleware
}

// NewChain 创建新的中间件链
func NewChain(middlewares ...Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
	}
}

// Add 添加中间件
func (c *Chain) Add(middleware Middleware) *Chain {
	c.middlewares = append(c.middlewares, middleware)
	return c
}

// Execute 执行中间件链
func (c *Chain) Execute(workflowName, nodeName string, nodeCtx core.Context, n node.Node, fn func() (interface{}, error)) (interface{}, error) {
	ctx := &Context{
		WorkflowName: workflowName,
		NodeName:     nodeName,
		StartTime:    time.Now(),
		Data:         make(map[string]interface{}),
	}

	// 执行 Before 中间件
	for _, m := range c.middlewares {
		if err := m.Before(ctx, nodeCtx, n); err != nil {
			return nil, err
		}
	}

	// 执行节点
	result, err := fn()

	// 执行 After 和 OnError 中间件
	for _, m := range c.middlewares {
		if err != nil {
			m.OnError(ctx, nodeCtx, n, err)
		} else {
			m.After(ctx, nodeCtx, n, result, nil)
		}
	}

	return result, err
}

// LoggingMiddleware 日志中间件
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Before 在节点执行前调用
func (m *LoggingMiddleware) Before(ctx *Context, nodeCtx core.Context, n node.Node) error {
	m.logger.Printf("[%s] 开始执行节点: %s", ctx.WorkflowName, n.Name())
	return nil
}

// After 在节点执行后调用
func (m *LoggingMiddleware) After(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error) {
	duration := time.Since(ctx.StartTime)
	m.logger.Printf("[%s] 节点 %s 执行完成，耗时: %v", ctx.WorkflowName, n.Name(), duration)
}

// OnError 在节点执行出错时调用
func (m *LoggingMiddleware) OnError(ctx *Context, nodeCtx core.Context, n node.Node, err error) {
	m.logger.Printf("[%s] 节点 %s 执行失败: %v", ctx.WorkflowName, n.Name(), err)
}

// MetricsMiddleware 指标中间件
type MetricsMiddleware struct {
	// TotalRequests 总请求数
	TotalRequests int64
	// SuccessCount 成功数
	SuccessCount int64
	// ErrorCount 失败数
	ErrorCount int64
	// TotalDuration 总耗时
	TotalDuration time.Duration
}

// NewMetricsMiddleware 创建指标中间件
func NewMetricsMiddleware() *MetricsMiddleware {
	return &MetricsMiddleware{}
}

// Before 在节点执行前调用
func (m *MetricsMiddleware) Before(ctx *Context, nodeCtx core.Context, n node.Node) error {
	m.TotalRequests++
	return nil
}

// After 在节点执行后调用
func (m *MetricsMiddleware) After(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error) {
	m.SuccessCount++
	m.TotalDuration += time.Since(ctx.StartTime)
}

// OnError 在节点执行出错时调用
func (m *MetricsMiddleware) OnError(ctx *Context, nodeCtx core.Context, n node.Node, err error) {
	m.ErrorCount++
	m.TotalDuration += time.Since(ctx.StartTime)
}

// RecoveryMiddleware 恢复中间件
type RecoveryMiddleware struct {
	handler func(ctx *Context, nodeCtx core.Context, n node.Node, r interface{})
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware(handler func(ctx *Context, nodeCtx core.Context, n node.Node, r interface{})) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		handler: handler,
	}
}

// Before 在节点执行前调用
func (m *RecoveryMiddleware) Before(ctx *Context, nodeCtx core.Context, n node.Node) error {
	return nil
}

// After 在节点执行后调用
func (m *RecoveryMiddleware) After(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error) {
}

// OnError 在节点执行出错时调用
func (m *RecoveryMiddleware) OnError(ctx *Context, nodeCtx core.Context, n node.Node, err error) {
	if m.handler != nil {
		m.handler(ctx, nodeCtx, n, err)
	}
}

// AuditMiddleware 审计中间件
type AuditMiddleware struct {
	handler func(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error, duration time.Duration)
}

// NewAuditMiddleware 创建审计中间件
func NewAuditMiddleware(handler func(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error, duration time.Duration)) *AuditMiddleware {
	return &AuditMiddleware{
		handler: handler,
	}
}

// Before 在节点执行前调用
func (m *AuditMiddleware) Before(ctx *Context, nodeCtx core.Context, n node.Node) error {
	return nil
}

// After 在节点执行后调用
func (m *AuditMiddleware) After(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error) {
	if m.handler != nil {
		m.handler(ctx, nodeCtx, n, result, nil, time.Since(ctx.StartTime))
	}
}

// OnError 在节点执行出错时调用
func (m *AuditMiddleware) OnError(ctx *Context, nodeCtx core.Context, n node.Node, err error) {
	if m.handler != nil {
		m.handler(ctx, nodeCtx, n, nil, err, time.Since(ctx.StartTime))
	}
}

// TracingMiddleware 追踪中间件
type TracingMiddleware struct {
	handler func(ctx *Context, nodeCtx core.Context, n node.Node, phase string, duration time.Duration)
}

// NewTracingMiddleware 创建追踪中间件
func NewTracingMiddleware(handler func(ctx *Context, nodeCtx core.Context, n node.Node, phase string, duration time.Duration)) *TracingMiddleware {
	return &TracingMiddleware{
		handler: handler,
	}
}

// Before 在节点执行前调用
func (m *TracingMiddleware) Before(ctx *Context, nodeCtx core.Context, n node.Node) error {
	if m.handler != nil {
		m.handler(ctx, nodeCtx, n, "start", 0)
	}
	return nil
}

// After 在节点执行后调用
func (m *TracingMiddleware) After(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error) {
	if m.handler != nil {
		m.handler(ctx, nodeCtx, n, "end", time.Since(ctx.StartTime))
	}
}

// OnError 在节点执行出错时调用
func (m *TracingMiddleware) OnError(ctx *Context, nodeCtx core.Context, n node.Node, err error) {
	if m.handler != nil {
		m.handler(ctx, nodeCtx, n, "error", time.Since(ctx.StartTime))
	}
}

// DefaultLoggingMiddleware 返回默认的日志中间件
func DefaultLoggingMiddleware() *LoggingMiddleware {
	return NewLoggingMiddleware(log.Default())
}

// String 返回中间件的字符串表示
func (c *Chain) String() string {
	return fmt.Sprintf("Chain{middlewares: %d}", len(c.middlewares))
}
