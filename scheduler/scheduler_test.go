package scheduler

import (
	"fmt"
	"testing"

	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/graph"
	"github.com/ELSTHub/elstworkflow/node"
)

func createTestNode(name string) node.Node {
	return node.NewFuncNode(name, func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})
}

func createTestGraph(t *testing.T) graph.Graph {
	t.Helper()
	g := graph.New()

	nodes := []string{"node1", "node2", "node3", "node4", "node5"}
	for _, name := range nodes {
		if err := g.AddNode(createTestNode(name)); err != nil {
			t.Fatalf("failed to add node %s: %v", name, err)
		}
	}

	// node1 -> node2 -> node4
	// node1 -> node3 -> node5
	// node4 -> node5
	edges := []struct{ from, to string }{
		{"node1", "node2"},
		{"node1", "node3"},
		{"node2", "node4"},
		{"node3", "node5"},
		{"node4", "node5"},
	}

	for _, e := range edges {
		if err := g.AddEdge(e.from, e.to); err != nil {
			t.Fatalf("failed to add edge %s -> %s: %v", e.from, e.to, err)
		}
	}

	return g
}

func TestNewSerialScheduler(t *testing.T) {
	g := createTestGraph(t)
	s, err := NewSerialScheduler(g)
	if err != nil {
		t.Fatalf("NewSerialScheduler failed: %v", err)
	}
	if s == nil {
		t.Fatal("scheduler is nil")
	}
}

func TestSerialSchedulerNext(t *testing.T) {
	g := createTestGraph(t)
	s, err := NewSerialScheduler(g)
	if err != nil {
		t.Fatalf("NewSerialScheduler failed: %v", err)
	}

	// 第一个应该是node1
	n, ok := s.Next()
	if !ok {
		t.Fatal("expected a node")
	}
	if n.Name() != "node1" {
		t.Errorf("expected 'node1', got '%s'", n.Name())
	}
}

func TestSerialSchedulerSchedule(t *testing.T) {
	g := createTestGraph(t)
	s, err := NewSerialScheduler(g)
	if err != nil {
		t.Fatalf("NewSerialScheduler failed: %v", err)
	}

	// 初始状态应该只有node1可调度
	runnable := s.Schedule()
	if len(runnable) != 1 {
		t.Fatalf("expected 1 runnable node, got %d", len(runnable))
	}
	if runnable[0].Name() != "node1" {
		t.Errorf("expected 'node1', got '%s'", runnable[0].Name())
	}
}

func TestSerialSchedulerMarkCompleted(t *testing.T) {
	g := createTestGraph(t)
	s, err := NewSerialScheduler(g)
	if err != nil {
		t.Fatalf("NewSerialScheduler failed: %v", err)
	}

	// 完成node1后，node2和node3应该可调度
	s.MarkCompleted("node1")

	runnable := s.Schedule()
	if len(runnable) != 2 {
		t.Fatalf("expected 2 runnable nodes, got %d", len(runnable))
	}

	names := make(map[string]bool)
	for _, n := range runnable {
		names[n.Name()] = true
	}

	if !names["node2"] {
		t.Error("expected node2 to be runnable")
	}
	if !names["node3"] {
		t.Error("expected node3 to be runnable")
	}
}

func TestSerialSchedulerIsCompleted(t *testing.T) {
	g := createTestGraph(t)
	s, err := NewSerialScheduler(g)
	if err != nil {
		t.Fatalf("NewSerialScheduler failed: %v", err)
	}

	if s.IsCompleted("node1") {
		t.Error("node1 should not be completed")
	}

	s.MarkCompleted("node1")

	if !s.IsCompleted("node1") {
		t.Error("node1 should be completed")
	}
}

func TestSerialSchedulerMarkFailed(t *testing.T) {
	g := createTestGraph(t)
	s, err := NewSerialScheduler(g)
	if err != nil {
		t.Fatalf("NewSerialScheduler failed: %v", err)
	}

	s.MarkFailed("node1")

	if !s.IsFailed("node1") {
		t.Error("node1 should be failed")
	}

	// 失败的节点不应该再被调度
	_, ok := s.Next()
	if ok {
		t.Error("should not have any runnable node")
	}
}

func TestSerialSchedulerRemaining(t *testing.T) {
	g := createTestGraph(t)
	s, err := NewSerialScheduler(g)
	if err != nil {
		t.Fatalf("NewSerialScheduler failed: %v", err)
	}

	if s.Remaining() != 5 {
		t.Errorf("expected 5 remaining, got %d", s.Remaining())
	}

	s.MarkCompleted("node1")
	if s.Remaining() != 4 {
		t.Errorf("expected 4 remaining, got %d", s.Remaining())
	}
}

func TestSerialSchedulerReset(t *testing.T) {
	g := createTestGraph(t)
	s, err := NewSerialScheduler(g)
	if err != nil {
		t.Fatalf("NewSerialScheduler failed: %v", err)
	}

	s.MarkCompleted("node1")
	s.MarkFailed("node2")

	s.Reset()

	if s.Remaining() != 5 {
		t.Errorf("expected 5 remaining after reset, got %d", s.Remaining())
	}
	if s.IsCompleted("node1") {
		t.Error("node1 should not be completed after reset")
	}
	if s.IsFailed("node2") {
		t.Error("node2 should not be failed after reset")
	}
}

func TestNewDAGScheduler(t *testing.T) {
	g := createTestGraph(t)
	s := NewDAGScheduler(g)
	if s == nil {
		t.Fatal("scheduler is nil")
	}
}

func TestDAGSchedulerNext(t *testing.T) {
	g := createTestGraph(t)
	s := NewDAGScheduler(g)

	n, ok := s.Next()
	if !ok {
		t.Fatal("expected a node")
	}
	if n.Name() != "node1" {
		t.Errorf("expected 'node1', got '%s'", n.Name())
	}
}

func TestDAGSchedulerSchedule(t *testing.T) {
	g := createTestGraph(t)
	s := NewDAGScheduler(g)

	runnable := s.Schedule()
	if len(runnable) != 1 {
		t.Fatalf("expected 1 runnable node, got %d", len(runnable))
	}
	if runnable[0].Name() != "node1" {
		t.Errorf("expected 'node1', got '%s'", runnable[0].Name())
	}
}

func TestDAGSchedulerMarkCompleted(t *testing.T) {
	g := createTestGraph(t)
	s := NewDAGScheduler(g)

	s.MarkCompleted("node1")

	runnable := s.Schedule()
	if len(runnable) != 2 {
		t.Fatalf("expected 2 runnable nodes, got %d", len(runnable))
	}

	names := make(map[string]bool)
	for _, n := range runnable {
		names[n.Name()] = true
	}

	if !names["node2"] {
		t.Error("expected node2 to be runnable")
	}
	if !names["node3"] {
		t.Error("expected node3 to be runnable")
	}
}

func TestDAGSchedulerFullExecution(t *testing.T) {
	g := createTestGraph(t)
	s := NewDAGScheduler(g)

	execOrder := make([]string, 0)

	for {
		n, ok := s.Next()
		if !ok {
			break
		}
		execOrder = append(execOrder, n.Name())
		s.MarkCompleted(n.Name())
	}

	if len(execOrder) != 5 {
		t.Fatalf("expected 5 nodes executed, got %d", len(execOrder))
	}

	// 验证拓扑顺序
	if execOrder[0] != "node1" {
		t.Errorf("first node should be 'node1', got '%s'", execOrder[0])
	}
	if execOrder[4] != "node5" {
		t.Errorf("last node should be 'node5', got '%s'", execOrder[4])
	}
}

func TestNewPriorityScheduler(t *testing.T) {
	g := createTestGraph(t)
	priorities := map[string]int{
		"node1": 1,
		"node2": 2,
		"node3": 3,
	}
	s := NewPriorityScheduler(g, priorities)
	if s == nil {
		t.Fatal("scheduler is nil")
	}
}

func TestPrioritySchedulerNext(t *testing.T) {
	g := createTestGraph(t)
	priorities := map[string]int{
		"node2": 1, // 更高优先级（数字越小优先级越高）
		"node3": 2,
	}
	s := NewPriorityScheduler(g, priorities)

	// 完成node1后，node2和node3都可运行，但node2优先级更高
	s.MarkCompleted("node1")

	n, ok := s.Next()
	if !ok {
		t.Fatal("expected a node")
	}
	if n.Name() != "node2" {
		t.Errorf("expected 'node2' (higher priority), got '%s'", n.Name())
	}
}

func TestSchedulerInterface(t *testing.T) {
	g := createTestGraph(t)

	// 确保所有调度器类型都实现了Scheduler接口
	s1, _ := NewSerialScheduler(g)
	var _ Scheduler = s1
	var _ Scheduler = NewDAGScheduler(g)
	var _ Scheduler = NewPriorityScheduler(g, nil)
}

// 基准测试
func BenchmarkSerialSchedulerNext(b *testing.B) {
	g := graph.New()
	for i := 0; i < 100; i++ {
		g.AddNode(createTestNode(fmt.Sprintf("node%d", i)))
	}
	s, _ := NewSerialScheduler(g)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Next()
	}
}

func BenchmarkDAGSchedulerSchedule(b *testing.B) {
	g := graph.New()
	for i := 0; i < 100; i++ {
		g.AddNode(createTestNode(fmt.Sprintf("node%d", i)))
	}
	s := NewDAGScheduler(g)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s.Schedule()
	}
}
