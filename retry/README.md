# Retry Package

retry 包提供了重试策略的实现，支持固定间隔、指数退避和自定义重试策略。

## 主要接口

### Retryer 接口

```go
type Retryer interface {
    NextInterval(attempt int) time.Duration
    ShouldRetry(err error, attempt int) bool
    MaxAttempts() int
    Reset()
}
```

## 重试策略

### 固定间隔重试

每次重试使用相同的间隔时间。

```go
policy := &core.RetryPolicy{
    Strategy:   core.RetryFixed,
    MaxRetries: 3,
    Interval:   time.Second,
}

retryer := retry.NewFixedRetryer(policy)
```

### 指数退避重试

每次重试间隔按指数增长。

```go
policy := &core.RetryPolicy{
    Strategy:    core.RetryExponential,
    MaxRetries:  5,
    Interval:    time.Second,
    Multiplier:  2.0,
    MaxInterval: 30 * time.Second,
}

retryer := retry.NewExponentialRetryer(policy)
```

### 自定义重试

使用自定义函数计算重试间隔。

```go
policy := &core.RetryPolicy{
    Strategy:   core.RetryCustom,
    MaxRetries: 3,
    Interval:   time.Second,
}

intervalFunc := func(attempt int) time.Duration {
    return time.Duration(attempt+1) * 500 * time.Millisecond
}

retryer := retry.NewCustomRetryer(policy, intervalFunc)
```

## 使用示例

### 基本使用

```go
retryer := retry.NewRetryer(policy)

result, err := retry.Execute(retryer, func() (interface{}, error) {
    // 执行可能失败的操作
    return doSomething()
})
```

### 带回调的重试

```go
result, err := retry.ExecuteWithCallback(retryer,
    func() (interface{}, error) {
        return doSomething()
    },
    func(attempt int, err error, nextInterval time.Duration) {
        fmt.Printf("重试第 %d 次，错误: %v，等待: %v\n", attempt, err, nextInterval)
    },
)
```

### 自定义重试条件

```go
policy := &core.RetryPolicy{
    Strategy:   core.RetryFixed,
    MaxRetries: 3,
    Interval:   time.Second,
    Retryable: func(err error) bool {
        // 只重试特定错误
        return errors.Is(err, ErrTemporary)
    },
}
```

## 设计原则

1. **接口统一**: 所有重试策略都实现 Retryer 接口
2. **可配置**: 支持多种重试策略和参数
3. **可扩展**: 支持自定义重试策略
4. **无状态**: 重试器本身不保存状态（除了指数退避的当前间隔）

## 重试策略选择

- **固定间隔**: 适用于简单的重试场景
- **指数退避**: 适用于需要避免重试风暴的场景
- **自定义**: 适用于需要特殊重试逻辑的场景

## 最佳实践

1. **设置最大重试次数**: 避免无限重试
2. **设置最大间隔**: 避免指数退避间隔过长
3. **使用重试条件**: 只重试可恢复的错误
4. **记录重试日志**: 便于问题排查
