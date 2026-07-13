# Core Package

core 包定义了工作流框架的核心类型和接口，是整个框架的基础。

## 主要类型

### 状态枚举

- `WorkflowStatus`: 工作流执行状态（Pending, Running, Completed, Failed, Cancelled, Paused）
- `NodeStatus`: 节点执行状态（Pending, Running, Completed, Failed, Skipped, Cancelled, RollingBack, RolledBack）
- `RetryStrategy`: 重试策略类型（Fixed, Exponential, Custom）
- `TimeoutAction`: 超时动作（Cancel, Fail, Retry）

### 策略类型

- `RetryPolicy`: 重试策略配置
- `TimeoutPolicy`: 超时策略配置
- `Options`: 通用选项配置

### 结果类型

- `NodeResult`: 节点执行结果
- `WorkflowResult`: 工作流执行结果

### 函数类型

- `Condition`: 条件函数，用于条件分支
- `NodeFunc`: 节点执行函数
- `RollbackFunc`: 回滚函数

### 接口

- `Context`: 工作流上下文接口，提供线程安全的键值存储

## 设计原则

1. **无业务依赖**: 所有类型都是通用的，不包含任何业务特定字段
2. **线程安全**: Context 接口设计为线程安全
3. **可扩展**: 通过接口和函数类型支持自定义扩展
4. **类型安全**: 使用泛型和类型断言提供类型安全

## 使用示例

```go
// 创建重试策略
retryPolicy := &core.RetryPolicy{
    Strategy:   core.RetryExponential,
    MaxRetries: 3,
    Interval:   time.Second,
    Multiplier: 2.0,
}

// 创建超时策略
timeoutPolicy := &core.TimeoutPolicy{
    Timeout:   30 * time.Second,
    OnTimeout: core.TimeoutCancel,
}

// 使用条件函数
condition := func(ctx core.Context) (bool, error) {
    val, ok := ctx.Get("approved")
    if !ok {
        return false, nil
    }
    approved, ok := val.(bool)
    return approved, nil
}
```
