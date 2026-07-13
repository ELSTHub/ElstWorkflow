# Scheduler Package

scheduler 包提供了工作流调度器的实现，负责寻找可执行的节点，不负责执行节点。

## 主要接口

### Scheduler 接口

```go
type Scheduler interface {
    Next() (node.Node, bool)
    Schedule() []node.Node
    MarkCompleted(name string)
    MarkFailed(name string)
    IsCompleted(name string) bool
    IsFailed(name string) bool
    Remaining() int
    Reset()
}
```

## 调度器类型

### Serial Scheduler

串行调度器，按照拓扑顺序依次执行节点。

```go
s, err := scheduler.NewSerialScheduler(graph)
for {
    n, ok := s.Next()
    if !ok {
        break
    }
    // 执行节点
    s.MarkCompleted(n.Name())
}
```

### DAG Scheduler

DAG 调度器，支持并行执行没有依赖关系的节点。

```go
s := scheduler.NewDAGScheduler(graph)
for s.Remaining() > 0 {
    runnable := s.Schedule()
    // 并行执行 runnable 中的节点
    for _, n := range runnable {
        go func(n node.Node) {
            // 执行节点
            s.MarkCompleted(n.Name())
        }(n)
    }
}
```

### Priority Scheduler

优先级调度器，按照优先级选择可执行的节点。

```go
priorities := map[string]int{
    "node1": 1,  // 高优先级
    "node2": 2,  // 中优先级
    "node3": 3,  // 低优先级
}
s := scheduler.NewPriorityScheduler(graph, priorities)
```

## 设计原则

1. **职责单一**: 只负责调度，不负责执行
2. **接口统一**: 所有调度器都实现 Scheduler 接口
3. **线程安全**: 所有操作都使用锁保护
4. **可扩展**: 支持自定义调度策略

## 使用场景

- **串行执行**: 简单的顺序执行
- **并行执行**: 没有依赖关系的节点并行执行
- **优先级调度**: 按照优先级选择执行顺序
