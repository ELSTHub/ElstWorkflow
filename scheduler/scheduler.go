// Package scheduler 提供了工作流调度器的实现。
// 调度器负责寻找可执行的节点，不负责执行节点。
package scheduler

import (
	"sync"

	"github.com/elstworkflow/graph"
	"github.com/elstworkflow/node"
)

// Scheduler 定义调度器接口
type Scheduler interface {
	// Next 返回下一个可执行的节点
	Next() (node.Node, bool)
	// Schedule 调度所有可执行的节点
	Schedule() []node.Node
	// MarkCompleted 标记节点为已完成
	MarkCompleted(name string)
	// MarkFailed 标记节点为失败
	MarkFailed(name string)
	// IsCompleted 检查节点是否已完成
	IsCompleted(name string) bool
	// IsFailed 检查节点是否失败
	IsFailed(name string) bool
	// Remaining 返回剩余未完成的节点数
	Remaining() int
	// Reset 重置调度器
	Reset()
}

// serialScheduler 串行调度器
type serialScheduler struct {
	mu         sync.RWMutex
	graph      graph.Graph
	completed  map[string]bool
	failed     map[string]bool
	sorted     []node.Node
	currentIdx int
}

// NewSerialScheduler 创建串行调度器
func NewSerialScheduler(g graph.Graph) (Scheduler, error) {
	sorted, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	return &serialScheduler{
		graph:     g,
		completed: make(map[string]bool),
		failed:    make(map[string]bool),
		sorted:    sorted,
	}, nil
}

// Next 返回下一个可执行的节点
func (s *serialScheduler) Next() (node.Node, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, n := range s.sorted {
		name := n.Name()
		if s.completed[name] || s.failed[name] {
			continue
		}

		// 检查所有依赖是否已完成
		deps := s.graph.GetNodeDependencies(name)
		allDepsCompleted := true
		for _, dep := range deps {
			if !s.completed[dep] {
				allDepsCompleted = false
				break
			}
		}

		if allDepsCompleted {
			return n, true
		}
	}

	return nil, false
}

// Schedule 调度所有可执行的节点
func (s *serialScheduler) Schedule() []node.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runnable := make([]node.Node, 0)
	for _, n := range s.sorted {
		name := n.Name()
		if s.completed[name] || s.failed[name] {
			continue
		}

		// 检查所有依赖是否已完成
		deps := s.graph.GetNodeDependencies(name)
		allDepsCompleted := true
		for _, dep := range deps {
			if !s.completed[dep] {
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

// MarkCompleted 标记节点为已完成
func (s *serialScheduler) MarkCompleted(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completed[name] = true
}

// MarkFailed 标记节点为失败
func (s *serialScheduler) MarkFailed(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failed[name] = true
}

// IsCompleted 检查节点是否已完成
func (s *serialScheduler) IsCompleted(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.completed[name]
}

// IsFailed 检查节点是否失败
func (s *serialScheduler) IsFailed(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.failed[name]
}

// Remaining 返回剩余未完成的节点数
func (s *serialScheduler) Remaining() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.sorted) - len(s.completed) - len(s.failed)
}

// Reset 重置调度器
func (s *serialScheduler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completed = make(map[string]bool)
	s.failed = make(map[string]bool)
}

// dagScheduler DAG调度器
type dagScheduler struct {
	mu        sync.RWMutex
	graph     graph.Graph
	completed map[string]bool
	failed    map[string]bool
}

// NewDAGScheduler 创建DAG调度器
func NewDAGScheduler(g graph.Graph) Scheduler {
	return &dagScheduler{
		graph:     g,
		completed: make(map[string]bool),
		failed:    make(map[string]bool),
	}
}

// Next 返回下一个可执行的节点
func (s *dagScheduler) Next() (node.Node, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runnable := s.findRunnableNodes()
	if len(runnable) == 0 {
		return nil, false
	}
	return runnable[0], true
}

// Schedule 调度所有可执行的节点
func (s *dagScheduler) Schedule() []node.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.findRunnableNodes()
}

// findRunnableNodes 查找可运行的节点
func (s *dagScheduler) findRunnableNodes() []node.Node {
	merged := make(map[string]bool)
	for k, v := range s.completed {
		merged[k] = v
	}
	for k, v := range s.failed {
		merged[k] = v
	}

	return s.graph.FindRunnableNodes(merged)
}

// MarkCompleted 标记节点为已完成
func (s *dagScheduler) MarkCompleted(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completed[name] = true
}

// MarkFailed 标记节点为失败
func (s *dagScheduler) MarkFailed(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failed[name] = true
}

// IsCompleted 检查节点是否已完成
func (s *dagScheduler) IsCompleted(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.completed[name]
}

// IsFailed 检查节点是否失败
func (s *dagScheduler) IsFailed(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.failed[name]
}

// Remaining 返回剩余未完成的节点数
func (s *dagScheduler) Remaining() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.graph.NodeCount() - len(s.completed) - len(s.failed)
}

// Reset 重置调度器
func (s *dagScheduler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completed = make(map[string]bool)
	s.failed = make(map[string]bool)
}

// priorityScheduler 优先级调度器
type priorityScheduler struct {
	mu         sync.RWMutex
	graph      graph.Graph
	completed  map[string]bool
	failed     map[string]bool
	priorities map[string]int
}

// NewPriorityScheduler 创建优先级调度器
func NewPriorityScheduler(g graph.Graph, priorities map[string]int) Scheduler {
	return &priorityScheduler{
		graph:      g,
		completed:  make(map[string]bool),
		failed:     make(map[string]bool),
		priorities: priorities,
	}
}

// Next 返回下一个可执行的节点（按优先级）
func (s *priorityScheduler) Next() (node.Node, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	runnable := s.findRunnableNodes()
	if len(runnable) == 0 {
		return nil, false
	}

	// 选择优先级最高的节点
	best := runnable[0]
	bestPriority := s.getNodePriority(best.Name())

	for _, n := range runnable[1:] {
		p := s.getNodePriority(n.Name())
		if p < bestPriority {
			best = n
			bestPriority = p
		}
	}

	return best, true
}

// Schedule 调度所有可执行的节点
func (s *priorityScheduler) Schedule() []node.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.findRunnableNodes()
}

// findRunnableNodes 查找可运行的节点
func (s *priorityScheduler) findRunnableNodes() []node.Node {
	merged := make(map[string]bool)
	for k, v := range s.completed {
		merged[k] = v
	}
	for k, v := range s.failed {
		merged[k] = v
	}

	return s.graph.FindRunnableNodes(merged)
}

// getNodePriority 获取节点优先级
func (s *priorityScheduler) getNodePriority(name string) int {
	if p, ok := s.priorities[name]; ok {
		return p
	}
	return 0
}

// MarkCompleted 标记节点为已完成
func (s *priorityScheduler) MarkCompleted(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completed[name] = true
}

// MarkFailed 标记节点为失败
func (s *priorityScheduler) MarkFailed(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failed[name] = true
}

// IsCompleted 检查节点是否已完成
func (s *priorityScheduler) IsCompleted(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.completed[name]
}

// IsFailed 检查节点是否失败
func (s *priorityScheduler) IsFailed(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.failed[name]
}

// Remaining 返回剩余未完成的节点数
func (s *priorityScheduler) Remaining() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.graph.NodeCount() - len(s.completed) - len(s.failed)
}

// Reset 重置调度器
func (s *priorityScheduler) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.completed = make(map[string]bool)
	s.failed = make(map[string]bool)
}
