# Builder Package

builder 包提供了工作流构建器模式实现，使用 Builder 模式可以方便地构建复杂的工作流图。

## 主要接口

### Builder 接口

```go
type Builder interface {
    Node(name string, executeFn core.NodeFunc, opts ...node.Option) Builder
    NodeWithRollback(name string, executeFn core.NodeFunc, rollbackFn core.RollbackFunc, opts ...node.Option) Builder
    StaticNode(name string, value interface{}, opts ...node.Option) Builder
    DependsOn(from string, deps ...string) Builder
    Parallel(name string) Builder
    Condition(name string, condition core.Condition) Builder
    WithRetryPolicy(name string, policy *core.RetryPolicy) Builder
    WithTimeoutPolicy(name string, policy *core.TimeoutPolicy) Builder
    WithMetadata(name string, metadata core.Metadata) Builder
    WithVersion(version string) Builder
    WithDescription(description string) Builder
    WithWorkflowMetadata(metadata core.Metadata) Builder
    WithWorkflowOptions(options *core.Options) Builder
    WithDependencies(deps []Dependency) Builder
    Build() (*Workflow, error)
}
```

### Workflow 结构体

```go
type Workflow struct {
    Name        string
    Version     string
    Description string
    Metadata    core.Metadata
    Graph       graph.Graph
    Options     *core.Options
}
```

## 使用示例

### 简单工作流

```go
wf, err := builder.New("my-workflow").
    Node("step1", func(ctx core.Context) (interface{}, error) {
        return "result1", nil
    }).
    Node("step2", func(ctx core.Context) (interface{}, error) {
        return "result2", nil
    }).
    DependsOn("step2", "step1").
    Build()
```

### 带回滚的工作流

```go
wf, err := builder.New("saga-workflow").
    NodeWithRollback("create-order",
        func(ctx core.Context) (interface{}, error) {
            // 创建订单
            return orderId, nil
        },
        func(ctx core.Context) error {
            // 回滚：取消订单
            return nil
        },
    ).
    NodeWithRollback("process-payment",
        func(ctx core.Context) (interface{}, error) {
            // 处理支付
            return paymentId, nil
        },
        func(ctx core.Context) error {
            // 回滚：退款
            return nil
        },
    ).
    DependsOn("process-payment", "create-order").
    Build()
```

### 并行执行

```go
wf, err := builder.New("parallel-workflow").
    Node("prepare", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    Node("task1", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    Node("task2", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    Node("finalize", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    DependsOn("task1", "prepare").
    DependsOn("task2", "prepare").
    DependsOn("finalize", "task1", "task2").
    Parallel("task1").
    Parallel("task2").
    Build()
```

### 链式依赖

```go
deps := builder.Chain("step1", "step2", "step3", "step4")

wf, err := builder.New("chain-workflow").
    Node("step1", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    Node("step2", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    Node("step3", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    Node("step4", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    WithDependencies(deps).
    Build()
```

### 带条件的工作流

```go
wf, err := builder.New("conditional-workflow").
    Node("check", func(ctx core.Context) (interface{}, error) {
        return true, nil
    }).
    Node("action", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    DependsOn("action", "check").
    Condition("action", func(ctx core.Context) (bool, error) {
        result, _ := ctx.Get("check")
        return result.(bool), nil
    }).
    Build()
```

### 完整配置

```go
wf, err := builder.New("full-workflow").
    WithVersion("1.0.0").
    WithDescription("Complete workflow example").
    WithWorkflowMetadata(core.Metadata{
        "env": "production",
        "team": "platform",
    }).
    Node("step1", func(ctx core.Context) (interface{}, error) {
        return nil, nil
    }).
    WithRetryPolicy("step1", &core.RetryPolicy{
        Strategy:   core.RetryExponential,
        MaxRetries: 3,
    }).
    WithTimeoutPolicy("step1", &core.TimeoutPolicy{
        Timeout: 30 * time.Second,
    }).
    WithMetadata("step1", core.Metadata{
        "priority": "high",
    }).
    Build()
```

## 设计原则

1. **流式接口**: 支持链式调用，提高代码可读性
2. **错误收集**: 收集所有构建错误，在 Build() 时统一报告
3. **验证**: 自动验证图结构和节点配置
4. **灵活性**: 支持多种节点类型和配置选项

## 错误处理

构建器会收集所有错误，在调用 Build() 时返回第一个错误：

```go
wf, err := builder.New("workflow").
    Node("node1", nil). // 错误：executeFn 为 nil
    Node("node1", nil). // 错误：重复节点
    Build()

if err != nil {
    // 处理错误
}
```
