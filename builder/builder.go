// Package builder 提供了工作流构建器模式实现。
// 使用 Builder 模式可以方便地构建复杂的工作流图。
package builder

import (
	"fmt"

	"github.com/elstworkflow/core"
	"github.com/elstworkflow/graph"
	"github.com/elstworkflow/node"
)

// Workflow 表示构建完成的工作流
type Workflow struct {
	// Name 工作流名称
	Name string
	// Version 工作流版本
	Version string
	// Description 工作流描述
	Description string
	// Metadata 元数据
	Metadata core.Metadata
	// Graph 工作流图
	Graph graph.Graph
	// Options 工作流选项
	Options *core.Options
}

// Builder 定义工作流构建器接口
type Builder interface {
	// Node 添加节点
	Node(name string, executeFn core.NodeFunc, opts ...node.Option) Builder
	// NodeWithRollback 添加带回滚的节点
	NodeWithRollback(name string, executeFn core.NodeFunc, rollbackFn core.RollbackFunc, opts ...node.Option) Builder
	// StaticNode 添加静态节点
	StaticNode(name string, value interface{}, opts ...node.Option) Builder
	// DependsOn 添加依赖关系
	DependsOn(from string, deps ...string) Builder
	// Parallel 设置节点为并行执行
	Parallel(name string) Builder
	// Condition 设置节点执行条件
	Condition(name string, condition core.Condition) Builder
	// WithRetryPolicy 设置重试策略
	WithRetryPolicy(name string, policy *core.RetryPolicy) Builder
	// WithTimeoutPolicy 设置超时策略
	WithTimeoutPolicy(name string, policy *core.TimeoutPolicy) Builder
	// WithMetadata 设置元数据
	WithMetadata(name string, metadata core.Metadata) Builder
	// WithVersion 设置工作流版本
	WithVersion(version string) Builder
	// WithDescription 设置工作流描述
	WithDescription(description string) Builder
	// WithWorkflowMetadata 设置工作流元数据
	WithWorkflowMetadata(metadata core.Metadata) Builder
	// WithWorkflowOptions 设置工作流选项
	WithWorkflowOptions(options *core.Options) Builder
	// WithDependencies 批量添加依赖关系
	WithDependencies(deps []Dependency) Builder
	// Build 构建工作流
	Build() (*Workflow, error)
}

// workflowBuilder 是 Builder 接口的默认实现
type workflowBuilder struct {
	name        string
	version     string
	description string
	metadata    core.Metadata
	options     *core.Options
	graph       graph.Graph
	nodes       map[string]node.Node
	nodeOpts    map[string]*node.Options
	errors      []error
}

// New 创建新的工作流构建器
func New(name string) Builder {
	return &workflowBuilder{
		name:     name,
		metadata: make(core.Metadata),
		options: &core.Options{
			Metadata: make(core.Metadata),
		},
		graph:    graph.New(),
		nodes:    make(map[string]node.Node),
		nodeOpts: make(map[string]*node.Options),
	}
}

// Node 添加节点
func (b *workflowBuilder) Node(name string, executeFn core.NodeFunc, opts ...node.Option) Builder {
	if b.hasNode(name) {
		b.errors = append(b.errors, fmt.Errorf("node %s already exists", name))
		return b
	}

	n := node.NewFuncNode(name, executeFn, opts...)
	b.addNode(n)
	return b
}

// NodeWithRollback 添加带回滚的节点
func (b *workflowBuilder) NodeWithRollback(name string, executeFn core.NodeFunc, rollbackFn core.RollbackFunc, opts ...node.Option) Builder {
	if b.hasNode(name) {
		b.errors = append(b.errors, fmt.Errorf("node %s already exists", name))
		return b
	}

	n := node.NewFuncNodeWithRollback(name, executeFn, rollbackFn, opts...)
	b.addNode(n)
	return b
}

// StaticNode 添加静态节点
func (b *workflowBuilder) StaticNode(name string, value interface{}, opts ...node.Option) Builder {
	if b.hasNode(name) {
		b.errors = append(b.errors, fmt.Errorf("node %s already exists", name))
		return b
	}

	n := node.NewStaticNode(name, value, opts...)
	b.addNode(n)
	return b
}

// addNode 添加节点到构建器
func (b *workflowBuilder) addNode(n node.Node) {
	name := n.Name()
	b.nodes[name] = n
	b.nodeOpts[name] = n.Options()
	b.graph.AddNode(n)
}

// hasNode 检查节点是否存在
func (b *workflowBuilder) hasNode(name string) bool {
	_, exists := b.nodes[name]
	return exists
}

// DependsOn 添加依赖关系
func (b *workflowBuilder) DependsOn(from string, deps ...string) Builder {
	for _, dep := range deps {
		if !b.hasNode(from) {
			b.errors = append(b.errors, fmt.Errorf("node %s does not exist", from))
			continue
		}
		if !b.hasNode(dep) {
			b.errors = append(b.errors, fmt.Errorf("dependency node %s does not exist", dep))
			continue
		}

		if err := b.graph.AddEdge(dep, from); err != nil {
			b.errors = append(b.errors, fmt.Errorf("failed to add dependency from %s to %s: %v", dep, from, err))
		}
	}
	return b
}

// Parallel 设置节点为并行执行
func (b *workflowBuilder) Parallel(name string) Builder {
	if !b.hasNode(name) {
		b.errors = append(b.errors, fmt.Errorf("node %s does not exist", name))
		return b
	}

	opts := b.nodeOpts[name]
	opts.Parallel = true
	return b
}

// Condition 设置节点执行条件
func (b *workflowBuilder) Condition(name string, condition core.Condition) Builder {
	if !b.hasNode(name) {
		b.errors = append(b.errors, fmt.Errorf("node %s does not exist", name))
		return b
	}

	opts := b.nodeOpts[name]
	opts.Condition = condition
	return b
}

// WithRetryPolicy 设置重试策略
func (b *workflowBuilder) WithRetryPolicy(name string, policy *core.RetryPolicy) Builder {
	if !b.hasNode(name) {
		b.errors = append(b.errors, fmt.Errorf("node %s does not exist", name))
		return b
	}

	opts := b.nodeOpts[name]
	opts.RetryPolicy = policy
	return b
}

// WithTimeoutPolicy 设置超时策略
func (b *workflowBuilder) WithTimeoutPolicy(name string, policy *core.TimeoutPolicy) Builder {
	if !b.hasNode(name) {
		b.errors = append(b.errors, fmt.Errorf("node %s does not exist", name))
		return b
	}

	opts := b.nodeOpts[name]
	opts.TimeoutPolicy = policy
	return b
}

// WithMetadata 设置元数据
func (b *workflowBuilder) WithMetadata(name string, metadata core.Metadata) Builder {
	if !b.hasNode(name) {
		b.errors = append(b.errors, fmt.Errorf("node %s does not exist", name))
		return b
	}

	opts := b.nodeOpts[name]
	if opts.Metadata == nil {
		opts.Metadata = make(core.Metadata)
	}
	for k, v := range metadata {
		opts.Metadata[k] = v
	}
	return b
}

// Build 构建工作流
func (b *workflowBuilder) Build() (*Workflow, error) {
	// 检查是否有构建错误
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("build errors: %v", b.errors[0])
	}

	// 验证图
	if err := b.graph.Validate(); err != nil {
		return nil, fmt.Errorf("graph validation failed: %v", err)
	}

	// 验证所有节点
	for _, n := range b.nodes {
		if err := n.Validate(); err != nil {
			return nil, fmt.Errorf("node %s validation failed: %v", n.Name(), err)
		}
	}

	return &Workflow{
		Name:        b.name,
		Version:     b.version,
		Description: b.description,
		Metadata:    b.metadata,
		Graph:       b.graph,
		Options:     b.options,
	}, nil
}

// WithVersion 设置工作流版本
func (b *workflowBuilder) WithVersion(version string) Builder {
	b.version = version
	return b
}

// WithDescription 设置工作流描述
func (b *workflowBuilder) WithDescription(description string) Builder {
	b.description = description
	return b
}

// WithWorkflowMetadata 设置工作流元数据
func (b *workflowBuilder) WithWorkflowMetadata(metadata core.Metadata) Builder {
	for k, v := range metadata {
		b.metadata[k] = v
	}
	return b
}

// WithWorkflowOptions 设置工作流选项
func (b *workflowBuilder) WithWorkflowOptions(options *core.Options) Builder {
	b.options = options
	return b
}

// Chain 创建链式依赖（A -> B -> C -> ...）
func Chain(nodes ...string) []Dependency {
	deps := make([]Dependency, 0, len(nodes)-1)
	for i := 1; i < len(nodes); i++ {
		deps = append(deps, Dependency{
			From: nodes[i-1],
			To:   nodes[i],
		})
	}
	return deps
}

// Dependency 表示依赖关系
type Dependency struct {
	From string
	To   string
}

// WithDependencies 批量添加依赖关系
func (b *workflowBuilder) WithDependencies(deps []Dependency) Builder {
	for _, dep := range deps {
		b.DependsOn(dep.To, dep.From)
	}
	return b
}
