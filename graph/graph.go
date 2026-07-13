// Package graph 实现了有向无环图（DAG）数据结构。
// 用于表示工作流中节点之间的依赖关系。
package graph

import (
	"fmt"
	"sync"

	"github.com/elstworkflow/node"
)

// Edge 表示图中的一条边
type Edge struct {
	// From 源节点名称
	From string
	// To 目标节点名称
	To string
}

// Graph 定义图接口
type Graph interface {
	// AddNode 添加节点
	AddNode(n node.Node) error
	// GetNode 获取节点
	GetNode(name string) (node.Node, bool)
	// RemoveNode 移除节点
	RemoveNode(name string) error
	// AddEdge 添加边
	AddEdge(from, to string) error
	// GetNodes 获取所有节点
	GetNodes() []node.Node
	// GetEdges 获取所有边
	GetEdges() []Edge
	// GetNodeDependencies 获取节点的依赖
	GetNodeDependencies(name string) []string
	// GetNodeDependents 获取依赖该节点的节点
	GetNodeDependents(name string) []string
	// TopologicalSort 拓扑排序
	TopologicalSort() ([]node.Node, error)
	// HasCycle 检测是否有环
	HasCycle() bool
	// FindRunnableNodes 查找可运行的节点（没有未完成依赖的节点）
	FindRunnableNodes(completed map[string]bool) []node.Node
	// Clone 克隆图
	Clone() Graph
	// Validate 验证图
	Validate() error
	// NodeCount 获取节点数量
	NodeCount() int
	// EdgeCount 获取边数量
	EdgeCount() int
}

// dagGraph 是 Graph 接口的默认实现
type dagGraph struct {
	mu       sync.RWMutex
	nodes    map[string]node.Node
	edges    map[string]map[string]bool // from -> to -> true
	reverse  map[string]map[string]bool // to -> from -> true (反向索引)
	edgeList []Edge
}

// New 创建新的 DAG 图
func New() Graph {
	return &dagGraph{
		nodes:   make(map[string]node.Node),
		edges:   make(map[string]map[string]bool),
		reverse: make(map[string]map[string]bool),
	}
}

// AddNode 添加节点
func (g *dagGraph) AddNode(n node.Node) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	name := n.Name()
	if name == "" {
		return fmt.Errorf("node name cannot be empty")
	}

	if _, exists := g.nodes[name]; exists {
		return fmt.Errorf("node %s already exists", name)
	}

	g.nodes[name] = n
	g.edges[name] = make(map[string]bool)
	g.reverse[name] = make(map[string]bool)

	return nil
}

// GetNode 获取节点
func (g *dagGraph) GetNode(name string) (node.Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	n, ok := g.nodes[name]
	return n, ok
}

// RemoveNode 移除节点
func (g *dagGraph) RemoveNode(name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[name]; !exists {
		return fmt.Errorf("node %s does not exist", name)
	}

	// 删除从该节点出发的边
	for to := range g.edges[name] {
		delete(g.reverse[to], name)
		// 从边列表中删除
		g.removeEdgeFromList(name, to)
	}
	delete(g.edges, name)

	// 删除指向该节点的边
	for from := range g.reverse[name] {
		delete(g.edges[from], name)
		// 从边列表中删除
		g.removeEdgeFromList(from, name)
	}
	delete(g.reverse, name)

	// 删除节点
	delete(g.nodes, name)

	return nil
}

// removeEdgeFromList 从边列表中删除边（内部使用，需要锁）
func (g *dagGraph) removeEdgeFromList(from, to string) {
	for i, edge := range g.edgeList {
		if edge.From == from && edge.To == to {
			g.edgeList = append(g.edgeList[:i], g.edgeList[i+1:]...)
			return
		}
	}
}

// AddEdge 添加边
func (g *dagGraph) AddEdge(from, to string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 检查节点是否存在
	if _, exists := g.nodes[from]; !exists {
		return fmt.Errorf("source node %s does not exist", from)
	}
	if _, exists := g.nodes[to]; !exists {
		return fmt.Errorf("target node %s does not exist", to)
	}

	// 检查是否是自环
	if from == to {
		return fmt.Errorf("self-loop detected: %s", from)
	}

	// 检查边是否已存在
	if g.edges[from][to] {
		return fmt.Errorf("edge from %s to %s already exists", from, to)
	}

	// 添加边
	g.edges[from][to] = true
	g.reverse[to][from] = true
	g.edgeList = append(g.edgeList, Edge{From: from, To: to})

	// 检查是否形成环
	if g.hasCycleInternal() {
		// 回滚
		delete(g.edges[from], to)
		delete(g.reverse[to], from)
		g.removeEdgeFromList(from, to)
		return fmt.Errorf("adding edge from %s to %s would create a cycle", from, to)
	}

	return nil
}

// GetNodes 获取所有节点
func (g *dagGraph) GetNodes() []node.Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	nodes := make([]node.Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

// GetEdges 获取所有边
func (g *dagGraph) GetEdges() []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges := make([]Edge, len(g.edgeList))
	copy(edges, g.edgeList)
	return edges
}

// GetNodeDependencies 获取节点的依赖（指向该节点的节点）
func (g *dagGraph) GetNodeDependencies(name string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	deps := make([]string, 0, len(g.reverse[name]))
	for from := range g.reverse[name] {
		deps = append(deps, from)
	}
	return deps
}

// GetNodeDependents 获取依赖该节点的节点（该节点指向的节点）
func (g *dagGraph) GetNodeDependents(name string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	dependents := make([]string, 0, len(g.edges[name]))
	for to := range g.edges[name] {
		dependents = append(dependents, to)
	}
	return dependents
}

// TopologicalSort 拓扑排序
func (g *dagGraph) TopologicalSort() ([]node.Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 计算入度
	inDegree := make(map[string]int)
	for name := range g.nodes {
		inDegree[name] = 0
	}
	for _, edges := range g.edges {
		for to := range edges {
			inDegree[to]++
		}
	}

	// 找到所有入度为0的节点
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]node.Node, 0, len(g.nodes))

	for len(queue) > 0 {
		// 取出第一个节点
		current := queue[0]
		queue = queue[1:]

		result = append(result, g.nodes[current])

		// 减少相邻节点的入度
		for to := range g.edges[current] {
			inDegree[to]--
			if inDegree[to] == 0 {
				queue = append(queue, to)
			}
		}
	}

	// 检查是否所有节点都在结果中
	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("graph contains a cycle")
	}

	return result, nil
}

// HasCycle 检测是否有环
func (g *dagGraph) HasCycle() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.hasCycleInternal()
}

// hasCycleInternal 内部环检测（需要锁）
func (g *dagGraph) hasCycleInternal() bool {
	// 使用DFS检测环
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for name := range g.nodes {
		if !visited[name] {
			if g.hasCycleDFS(name, visited, recStack) {
				return true
			}
		}
	}

	return false
}

// hasCycleDFS DFS环检测
func (g *dagGraph) hasCycleDFS(node string, visited, recStack map[string]bool) bool {
	visited[node] = true
	recStack[node] = true

	for to := range g.edges[node] {
		if !visited[to] {
			if g.hasCycleDFS(to, visited, recStack) {
				return true
			}
		} else if recStack[to] {
			return true
		}
	}

	recStack[node] = false
	return false
}

// FindRunnableNodes 查找可运行的节点
func (g *dagGraph) FindRunnableNodes(completed map[string]bool) []node.Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	runnable := make([]node.Node, 0)

	for name, n := range g.nodes {
		// 跳过已完成的节点
		if completed[name] {
			continue
		}

		// 检查所有依赖是否已完成
		allDepsCompleted := true
		for from := range g.reverse[name] {
			if !completed[from] {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			runnable = append(runnable, n)
		}
	}

	return runnable
}

// Clone 克隆图
func (g *dagGraph) Clone() Graph {
	g.mu.RLock()
	defer g.mu.RUnlock()

	cloned := &dagGraph{
		nodes:    make(map[string]node.Node, len(g.nodes)),
		edges:    make(map[string]map[string]bool, len(g.edges)),
		reverse:  make(map[string]map[string]bool, len(g.reverse)),
		edgeList: make([]Edge, len(g.edgeList)),
	}

	// 复制节点
	for name, n := range g.nodes {
		cloned.nodes[name] = n
	}

	// 复制边
	for from, tos := range g.edges {
		cloned.edges[from] = make(map[string]bool, len(tos))
		for to := range tos {
			cloned.edges[from][to] = true
		}
	}

	// 复制反向索引
	for to, froms := range g.reverse {
		cloned.reverse[to] = make(map[string]bool, len(froms))
		for from := range froms {
			cloned.reverse[to][from] = true
		}
	}

	// 复制边列表
	copy(cloned.edgeList, g.edgeList)

	return cloned
}

// Validate 验证图
func (g *dagGraph) Validate() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 检查是否有环
	if g.hasCycleInternal() {
		return fmt.Errorf("graph contains a cycle")
	}

	// 验证所有节点
	for _, n := range g.nodes {
		if err := n.Validate(); err != nil {
			return fmt.Errorf("node %s validation failed: %v", n.Name(), err)
		}
	}

	return nil
}

// NodeCount 获取节点数量
func (g *dagGraph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.nodes)
}

// EdgeCount 获取边数量
func (g *dagGraph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.edgeList)
}
