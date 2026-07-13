# Rollback Package

rollback 包提供了 Saga 模式的回滚管理实现，支持回滚栈、部分补偿和跳过回滚等功能。

## 主要接口

### RollbackManager 接口

```go
type RollbackManager interface {
    Push(compensation Compensation)
    Pop() (Compensation, bool)
    Rollback(ctx core.Context) error
    RollbackUntil(nodeName string, ctx core.Context) error
    RollbackOnly(nodeName string, ctx core.Context) error
    SkipRollback(nodeName string)
    Clear()
    Size() int
    HasCompensation(nodeName string) bool
}
```

### Compensation 结构体

```go
type Compensation struct {
    NodeName     string
    RollbackFunc core.RollbackFunc
    SkipRollback bool
    RollbackOnly bool
    Context      core.Context
}
```

### RollbackError 结构体

```go
type RollbackError struct {
    NodeName string
    Err      error
    Partial  bool
}
```

## 使用示例

### 基本回滚

```go
manager := rollback.NewRollbackManager()

// 执行操作并注册回滚
manager.Push(rollback.Compensation{
    NodeName: "create-order",
    RollbackFunc: func(ctx core.Context) error {
        // 取消订单
        return cancelOrder(orderId)
    },
    Context: ctx,
})

// 如果后续操作失败，执行回滚
if err := doSomething(); err != nil {
    manager.Rollback(ctx)
}
```

### 部分回滚

```go
// 从指定节点开始回滚
err := manager.RollbackUntil("create-order", ctx)
if err != nil {
    // 处理部分回滚失败
    var rollbackErr *rollback.RollbackError
    if errors.As(err, &rollbackErr) && rollbackErr.Partial {
        // 部分节点已回滚
    }
}
```

### 跳过回滚

```go
// 标记节点跳过回滚
manager.SkipRollback("read-only-node")

// 或在补偿操作中指定
manager.Push(rollback.Compensation{
    NodeName:     "read-only-node",
    RollbackFunc: nil,
    SkipRollback: true,
    Context:      ctx,
})
```

### 只执行回滚

```go
// 只执行指定节点的回滚
err := manager.RollbackOnly("create-order", ctx)
```

## 设计原则

1. **栈结构**: 后进先出的回滚顺序
2. **部分补偿**: 支持部分节点回滚成功的情况
3. **跳过机制**: 支持跳过不需要回滚的节点
4. **线程安全**: 使用互斥锁保证并发安全

## 回滚顺序

回滚按照后进先出（LIFO）的顺序执行：

```
执行顺序: node1 -> node2 -> node3
回滚顺序: node3 -> node2 -> node1
```

## 错误处理

当回滚过程中发生错误时：

1. 继续执行剩余节点的回滚
2. 返回最后一个错误
3. 标记是否为部分回滚

```go
err := manager.Rollback(ctx)
if err != nil {
    var rollbackErr *rollback.RollbackError
    if errors.As(err, &rollbackErr) {
        fmt.Printf("节点: %s, 错误: %v, 部分回滚: %v\n",
            rollbackErr.NodeName, rollbackErr.Err, rollbackErr.Partial)
    }
}
```

## 最佳实践

1. **及时注册**: 在执行操作前注册回滚函数
2. **幂等回滚**: 回滚函数应该是幂等的
3. **处理错误**: 妥善处理回滚过程中的错误
4. **记录日志**: 记录回滚操作的日志
