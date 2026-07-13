# Engine Package

engine 包提供了工作流引擎的实现，是整个工作流的运行时，负责加载工作流、调度、执行、回滚等功能。

## 主要接口

### Engine 接口

```go
type Engine interface {
    Load(wf *builder.Workflow) error
    Run(ctx core.Context) (*core.WorkflowResult, error)
    Pause() error
    Resume(ctx core.Context) (*core.WorkflowResult, error)
    Cancel() error
    Status() Status
    NodeResults() map[string]*core.NodeResult
}
```

### Config 结构体

```go
type Config struct {
    MaxParallel   int
    Executor      executor.Executor
    SchedulerType SchedulerType
}
```

## 使用示例

### 基本使用

```go
// 构建工作流
wf, _ := builder.New("my-workflow").
    Node("step1", func(ctx core.Context) (interface{}, error) {
        return "result1", nil
    }).
    Node("step2", func(ctx core.Context) (interface{}, error) {
        return "result2", nil
    }).
    DependsOn("step2", "step1").
    Build()

// 创建引擎
e := engine.New(nil)

// 加载工作流
e.Load(wf)

// 运行工作流
ctx := context.New()
result, err := e.Run(ctx)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("工作流状态: %v\n", result.Status)
```

### 使用 DAG 调度器

```go
config := &engine.Config{
    MaxParallel:   4,
    SchedulerType: engine.DAGScheduler,
}

e := engine.New(config)
```

### 便捷函数

```go
result, err := engine.RunSimple(wf, ctx)
```

## 引擎状态

- **Idle**: 空闲状态
- **Running**: 运行中
- **Paused**: 已暂停
- **Completed**: 已完成
- **Failed**: 失败
- **Cancelled**: 已取消

## 设计原则

1. **职责分离**: 引擎负责协调，不包含业务逻辑
2. **可配置**: 支持自定义执行器和调度器
3. **状态管理**: 完整的工作流状态管理
4. **线程安全**: 所有操作都使用锁保护

## 状态转换

```
Idle -> Running -> Completed
                 -> Failed
                 -> Cancelled
Running -> Paused -> Running
                  -> Cancelled
```
