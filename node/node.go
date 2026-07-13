// Package node 定义了工作流节点的接口和默认实现。
// 节点是工作流执行的基本单元，每个节点负责一个具体的任务。
package node

import (
	"fmt"
	"time"

	"github.com/ELSTHub/elstworkflow/core"
)

// Node 定义工作流节点接口
type Node interface {
	// Name 返回节点名称
	Name() string

	// Execute 执行节点任务
	Execute(ctx core.Context) (interface{}, error)

	// Rollback 回滚节点操作
	Rollback(ctx core.Context) error

	// Validate 验证节点配置
	Validate() error

	// Options 返回节点选项
	Options() *Options
}

// Options 定义节点选项
type Options struct {
	// Name 节点名称
	Name string
	// Description 节点描述
	Description string
	// RetryPolicy 重试策略
	RetryPolicy *core.RetryPolicy
	// TimeoutPolicy 超时策略
	TimeoutPolicy *core.TimeoutPolicy
	// Parallel 是否可并行执行
	Parallel bool
	// Metadata 元数据
	Metadata core.Metadata
	// Dependencies 依赖的节点名称
	Dependencies []string
	// Condition 执行条件
	Condition core.Condition
}

// Option 定义节点选项函数类型
type Option func(*Options)

// WithName 设置节点名称
func WithName(name string) Option {
	return func(o *Options) {
		o.Name = name
	}
}

// WithDescription 设置节点描述
func WithDescription(description string) Option {
	return func(o *Options) {
		o.Description = description
	}
}

// WithRetryPolicy 设置重试策略
func WithRetryPolicy(policy *core.RetryPolicy) Option {
	return func(o *Options) {
		o.RetryPolicy = policy
	}
}

// WithTimeoutPolicy 设置超时策略
func WithTimeoutPolicy(policy *core.TimeoutPolicy) Option {
	return func(o *Options) {
		o.TimeoutPolicy = policy
	}
}

// WithParallel 设置并行执行标志
func WithParallel(parallel bool) Option {
	return func(o *Options) {
		o.Parallel = parallel
	}
}

// WithMetadata 设置元数据
func WithMetadata(metadata core.Metadata) Option {
	return func(o *Options) {
		o.Metadata = metadata
	}
}

// WithDependencies 设置依赖节点
func WithDependencies(deps ...string) Option {
	return func(o *Options) {
		o.Dependencies = deps
	}
}

// WithCondition 设置执行条件
func WithCondition(condition core.Condition) Option {
	return func(o *Options) {
		o.Condition = condition
	}
}

// baseNode 提供节点的基础实现
type baseNode struct {
	options *Options
}

// newBaseNode 创建基础节点
func newBaseNode(opts ...Option) *baseNode {
	options := &Options{
		Metadata: make(core.Metadata),
	}

	for _, opt := range opts {
		opt(options)
	}

	return &baseNode{
		options: options,
	}
}

// Name 返回节点名称
func (n *baseNode) Name() string {
	return n.options.Name
}

// Options 返回节点选项
func (n *baseNode) Options() *Options {
	return n.options
}

// Validate 验证节点配置
func (n *baseNode) Validate() error {
	if n.options.Name == "" {
		return fmt.Errorf("node name is required")
	}
	return nil
}

// funcNode 是基于函数的节点实现
type funcNode struct {
	*baseNode
	executeFn  core.NodeFunc
	rollbackFn core.RollbackFunc
}

// NewFuncNode 创建基于函数的节点
func NewFuncNode(name string, executeFn core.NodeFunc, opts ...Option) Node {
	options := append([]Option{WithName(name)}, opts...)
	return &funcNode{
		baseNode:   newBaseNode(options...),
		executeFn:  executeFn,
		rollbackFn: nil,
	}
}

// NewFuncNodeWithRollback 创建带回滚函数的节点
func NewFuncNodeWithRollback(name string, executeFn core.NodeFunc, rollbackFn core.RollbackFunc, opts ...Option) Node {
	options := append([]Option{WithName(name)}, opts...)
	return &funcNode{
		baseNode:   newBaseNode(options...),
		executeFn:  executeFn,
		rollbackFn: rollbackFn,
	}
}

// Execute 执行节点任务
func (n *funcNode) Execute(ctx core.Context) (interface{}, error) {
	if n.executeFn == nil {
		return nil, fmt.Errorf("execute function is not defined for node %s", n.Name())
	}
	return n.executeFn(ctx)
}

// Rollback 回滚节点操作
func (n *funcNode) Rollback(ctx core.Context) error {
	if n.rollbackFn == nil {
		return nil // 没有回滚函数，认为回滚成功
	}
	return n.rollbackFn(ctx)
}

// Validate 验证节点配置
func (n *funcNode) Validate() error {
	if err := n.baseNode.Validate(); err != nil {
		return err
	}
	if n.executeFn == nil {
		return fmt.Errorf("execute function is required for node %s", n.Name())
	}
	return nil
}

// staticNode 是静态值节点，直接返回预设值
type staticNode struct {
	*baseNode
	value interface{}
}

// NewStaticNode 创建静态值节点
func NewStaticNode(name string, value interface{}, opts ...Option) Node {
	options := append([]Option{WithName(name)}, opts...)
	return &staticNode{
		baseNode: newBaseNode(options...),
		value:    value,
	}
}

// Execute 返回预设的静态值
func (n *staticNode) Execute(ctx core.Context) (interface{}, error) {
	return n.value, nil
}

// Rollback 静态节点无需回滚
func (n *staticNode) Rollback(ctx core.Context) error {
	return nil
}

// errorNode 总是返回错误的节点，用于测试
type errorNode struct {
	*baseNode
	err error
}

// NewErrorNode 创建总是返回错误的节点
func NewErrorNode(name string, err error, opts ...Option) Node {
	options := append([]Option{WithName(name)}, opts...)
	return &errorNode{
		baseNode: newBaseNode(options...),
		err:      err,
	}
}

// Execute 返回预设的错误
func (n *errorNode) Execute(ctx core.Context) (interface{}, error) {
	return nil, n.err
}

// Rollback 错误节点无需回滚
func (n *errorNode) Rollback(ctx core.Context) error {
	return nil
}

// delayNode 延迟指定时间后执行的节点
type delayNode struct {
	*baseNode
	delay     time.Duration
	executeFn core.NodeFunc
}

// NewDelayNode 创建延迟节点
func NewDelayNode(name string, delay time.Duration, executeFn core.NodeFunc, opts ...Option) Node {
	options := append([]Option{WithName(name)}, opts...)
	return &delayNode{
		baseNode:  newBaseNode(options...),
		delay:     delay,
		executeFn: executeFn,
	}
}

// Execute 延迟后执行节点任务
func (n *delayNode) Execute(ctx core.Context) (interface{}, error) {
	time.Sleep(n.delay)
	if n.executeFn == nil {
		return nil, nil
	}
	return n.executeFn(ctx)
}

// Rollback 延迟节点无需回滚
func (n *delayNode) Rollback(ctx core.Context) error {
	return nil
}
