# Timeout Package

timeout 包提供了超时策略的实现，支持工作流超时、节点超时和上下文取消。

## 主要接口

### TimeoutManager 接口

```go
type TimeoutManager interface {
    WithTimeout(timeout time.Duration, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
    WithTimeoutPolicy(policy *core.TimeoutPolicy, fn func(ctx context.Context) (interface{}, error)) (interface{}, error)
    CreateContext(timeout time.Duration) (context.Context, context.CancelFunc)
    CreateContextWithParent(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc)
}
```

### TimeoutError 结构体

```go
type TimeoutError struct {
    Operation string
    Timeout   time.Duration
    Elapsed   time.Duration
}
```

## 使用示例

### 基本超时

```go
manager := timeout.NewTimeoutManager()

result, err := manager.WithTimeout(5*time.Second, func(ctx context.Context) (interface{}, error) {
    // 执行可能耗时的操作
    return doSomething(ctx)
})
```

### 使用超时策略

```go
policy := &core.TimeoutPolicy{
    Timeout:   30 * time.Second,
    OnTimeout: core.TimeoutCancel,
}

result, err := manager.WithTimeoutPolicy(policy, func(ctx context.Context) (interface{}, error) {
    return doSomething(ctx)
})
```

### 创建超时上下文

```go
ctx, cancel := manager.CreateContext(10 * time.Second)
defer cancel()

// 使用上下文执行操作
select {
case result := <-doAsyncWork(ctx):
    // 处理结果
case <-ctx.Done():
    // 超时处理
}
```

### 包装函数

```go
// 包装现有函数，添加超时功能
fn := func() (interface{}, error) {
    return doSomething()
}

wrapped := timeout.WrapWithTimeout(fn, 5*time.Second)
result, err := wrapped()
```

### 检查超时错误

```go
_, err := timeout.ExecuteWithTimeout(1*time.Second, func(ctx context.Context) (interface{}, error) {
    time.Sleep(10 * time.Second)
    return nil, nil
})

if timeout.IsTimeoutError(err) {
    // 处理超时
}
```

## 设计原则

1. **上下文集成**: 基于 Go 标准库的 context 包
2. **策略模式**: 支持超时策略配置
3. **错误类型化**: 提供专门的超时错误类型
4. **可组合**: 支持函数包装和组合

## 超时策略

### 工作流超时

整个工作流的执行超时。

```go
policy := &core.TimeoutPolicy{
    Timeout:   5 * time.Minute,
    OnTimeout: core.TimeoutFail,
}
```

### 节点超时

单个节点的执行超时。

```go
policy := &core.TimeoutPolicy{
    Timeout:   30 * time.Second,
    OnTimeout: core.TimeoutRetry,
}
```

## 最佳实践

1. **合理设置超时**: 根据操作的实际耗时设置超时
2. **处理超时错误**: 优雅地处理超时情况
3. **使用上下文传递**: 通过上下文传递取消信号
4. **避免过短超时**: 避免设置过短的超时导致正常操作失败

## 与 context 包的关系

timeout 包基于 Go 标准库的 context 包实现，提供了更高级的超时管理功能：

- 简化超时上下文的创建
- 提供策略化的超时配置
- 提供专门的超时错误类型
- 支持函数包装和组合
