# Node Package

node 包定义了工作流节点的接口和多种默认实现，节点是工作流执行的基本单元。

## 主要接口

### Node 接口

```go
type Node interface {
    Name() string
    Execute(ctx core.Context) (interface{}, error)
    Rollback(ctx core.Context) error
    Validate() error
    Options() *Options
}
```

### Options 结构体

```go
type Options struct {
    Name         string
    Description  string
    RetryPolicy  *core.RetryPolicy
    TimeoutPolicy *core.TimeoutPolicy
    Parallel     bool
    Metadata     core.Metadata
    Dependencies []string
    Condition    core.Condition
}
```

## 节点类型

### FuncNode

基于函数的节点，最灵活的节点类型。

```go
node := node.NewFuncNode("my-node", func(ctx core.Context) (interface{}, error) {
    // 执行业务逻辑
    return result, nil
})
```

### FuncNodeWithRollback

带回滚函数的节点，支持 Saga 模式。

```go
node := node.NewFuncNodeWithRollback("my-node",
    func(ctx core.Context) (interface{}, error) {
        // 执行操作
        return result, nil
    },
    func(ctx core.Context) error {
        // 回滚操作
        return nil
    },
)
```

### StaticNode

静态值节点，直接返回预设值。

```go
node := node.NewStaticNode("config", map[string]string{
    "key": "value",
})
```

### ErrorNode

总是返回错误的节点，主要用于测试。

```go
node := node.NewErrorNode("fail", errors.New("something went wrong"))
```

### DelayNode

延迟指定时间后执行的节点。

```go
node := node.NewDelayNode("delay", 5*time.Second, func(ctx core.Context) (interface{}, error) {
    return "done", nil
})
```

## 选项函数

使用函数式选项模式配置节点：

```go
node := node.NewFuncNode("my-node", executeFn,
    node.WithDescription("My custom node"),
    node.WithRetryPolicy(&core.RetryPolicy{
        Strategy:   core.RetryExponential,
        MaxRetries: 3,
    }),
    node.WithTimeoutPolicy(&core.TimeoutPolicy{
        Timeout: 30 * time.Second,
    }),
    node.WithParallel(true),
    node.WithMetadata(core.Metadata{"env": "production"}),
    node.WithDependencies("node1", "node2"),
    node.WithCondition(func(ctx core.Context) (bool, error) {
        return true, nil
    }),
)
```

## 设计原则

1. **接口统一**: 所有节点类型都实现 Node 接口
2. **可组合**: 通过选项函数灵活配置
3. **可验证**: 支持 Validate() 验证配置
4. **可回滚**: 支持 Saga 模式的回滚操作
5. **无业务依赖**: 节点不包含任何业务特定逻辑

## 扩展节点

实现自定义节点类型：

```go
type CustomNode struct {
    *node.baseNode
    // 自定义字段
}

func (n *CustomNode) Execute(ctx core.Context) (interface{}, error) {
    // 自定义执行逻辑
    return nil, nil
}

func (n *CustomNode) Rollback(ctx core.Context) error {
    // 自定义回滚逻辑
    return nil
}
```
