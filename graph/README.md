# Graph Package

graph 包实现了有向无环图（DAG）数据结构，用于表示工作流中节点之间的依赖关系。

## 主要接口

### Graph 接口

```go
type Graph interface {
    AddNode(n node.Node) error
    GetNode(name string) (node.Node, bool)
    RemoveNode(name string) error
    AddEdge(from, to string) error
    GetNodes() []node.Node
    GetEdges() []Edge
    GetNodeDependencies(name string) []string
    GetNodeDependents(name string) []string
    TopologicalSort() ([]node.Node, error)
    HasCycle() bool
    FindRunnableNodes(completed map[string]bool) []node.Node
    Clone() Graph
    Validate() error
    NodeCount() int
    EdgeCount() int
}
```

### Edge 结构体

```go
type Edge struct {
    From string
    To   string
}
```

## 主要功能

### 节点管理

```go
g := graph.New()

// 添加节点
g.AddNode(node.NewFuncNode("node1", executeFn))
g.AddNode(node.NewFuncNode("node2", executeFn))

// 获取节点
n, ok := g.GetNode("node1")

// 移除节点
g.RemoveNode("node1")
```

### 边管理

```go
// 添加边（node1 -> node2）
g.AddEdge("node1", "node2")

// 获取所有边
edges := g.GetEdges()

// 获取节点依赖
deps := g.GetNodeDependencies("node2") // ["node1"]

// 获取依赖该节点的节点
dependents := g.GetNodeDependents("node1") // ["node2"]
```

### 拓扑排序

```go
sorted, err := g.TopologicalSort()
if err != nil {
    // 图中有环
}

for _, n := range sorted {
    fmt.Println(n.Name())
}
```

### 环检测

```go
if g.HasCycle() {
    // 图中有环
}
```

### 查找可运行节点

```go
completed := map[string]bool{
    "node1": true,
    "node2": true,
}

runnable := g.FindRunnableNodes(completed)
// 返回所有依赖都已完成的节点
```

### 图克隆

```go
cloned := g.Clone()

// 修改原始不影响克隆
// 修改克隆不影响原始
```

### 图验证

```go
if err := g.Validate(); err != nil {
    // 图无效（有环或节点验证失败）
}
```

## 设计原则

1. **线程安全**: 所有操作都使用读写锁保护
2. **DAG 保证**: 添加边时自动检测环，保证图始终是 DAG
3. **接口隔离**: 实现 Graph 接口，不暴露内部实现
4. **可克隆**: 支持深拷贝，修改克隆不影响原始

## 使用场景

1. **工作流调度**: 表示任务之间的依赖关系
2. **构建系统**: 表示编译顺序
3. **数据处理**: 表示数据处理管道
4. **任务编排**: 表示任务执行顺序

## 性能考虑

- 节点和边的查找是 O(1) 操作
- 拓扑排序是 O(V + E) 操作
- 环检测是 O(V + E) 操作
- 克隆是 O(V + E) 操作
