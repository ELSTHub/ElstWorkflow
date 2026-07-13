# Executor Package

executor 包提供了节点执行器的实现，负责执行节点，处理重试、超时和中间件。

## 主要接口

### Executor 接口

```go
type Executor interface {
    Execute(ctx core.Context, n node.Node) (*core.NodeResult, error)
    ExecuteWithRollback(ctx core.Context, n node.Node, rollbackMgr rollback.RollbackManager) (*core.NodeResult, error)
}
```

### Middleware 类型

```go
type Middleware func(next ExecuteFunc) ExecuteFunc
type ExecuteFunc func(ctx core.Context, n node.Node) (*core.NodeResult, error)
```

## 使用示例

### 基本执行

```go
e := executor.New()
ctx := context.New()

n := node.NewFuncNode("my-node", func(ctx core.Context) (interface{}, error) {
    return "result", nil
})

result, err := e.Execute(ctx, n)
```

### 使用中间件

```go
loggingMiddleware := func(next executor.ExecuteFunc) executor.ExecuteFunc {
    return func(ctx core.Context, n node.Node) (*core.NodeResult, error) {
        log.Printf("开始执行节点: %s", n.Name())
        result, err := next(ctx, n)
        log.Printf("节点 %s 执行完成", n.Name())
        return result, err
    }
}

e := executor.New(executor.WithMiddleware(loggingMiddleware))
```

### 使用重试

```go
retryer := retry.NewFixedRetryer(&core.RetryPolicy{
    Strategy:   core.RetryFixed,
    MaxRetries: 3,
    Interval:   time.Second,
})

e := executor.New(executor.WithRetryer(retryer))
```

### 带回滚的执行

```go
rollbackMgr := rollback.NewRollbackManager()

n := node.NewFuncNodeWithRollback("my-node",
    func(ctx core.Context) (interface{}, error) {
        return doSomething()
    },
    func(ctx core.Context) error {
        return undoSomething()
    },
)

result, err := e.ExecuteWithRollback(ctx, n, rollbackMgr)
```

## 设计原则

1. **职责单一**: 只负责执行，不负责调度
2. **中间件模式**: 支持可组合的中间件链
3. **可配置**: 支持重试和超时策略
4. **线程安全**: 支持并发执行

## 中间件示例

### 日志中间件

```go
func LoggingMiddleware() executor.Middleware {
    return func(next executor.ExecuteFunc) executor.ExecuteFunc {
        return func(ctx core.Context, n node.Node) (*core.NodeResult, error) {
            start := time.Now()
            result, err := next(ctx, n)
            log.Printf("节点 %s 执行耗时: %v", n.Name(), time.Since(start))
            return result, err
        }
    }
}
```

### 指标中间件

```go
func MetricsMiddleware() executor.Middleware {
    return func(next executor.ExecuteFunc) executor.ExecuteFunc {
        return func(ctx core.Context, n node.Node) (*core.NodeResult, error) {
            // 记录指标
            return next(ctx, n)
        }
    }
}
```
