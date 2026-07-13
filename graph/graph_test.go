package graph

import (
	"fmt"
	"testing"

	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/node"
)

func createTestNode(name string) node.Node {
	return node.NewFuncNode(name, func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})
}

func TestNewGraph(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.NodeCount() != 0 {
		t.Errorf("expected 0 nodes, got %d", g.NodeCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("expected 0 edges, got %d", g.EdgeCount())
	}
}

func TestAddNode(t *testing.T) {
	g := New()
	n := createTestNode("node1")

	if err := g.AddNode(n); err != nil {
		t.Errorf("AddNode failed: %v", err)
	}

	if g.NodeCount() != 1 {
		t.Errorf("expected 1 node, got %d", g.NodeCount())
	}

	// 测试获取节点
	retrieved, ok := g.GetNode("node1")
	if !ok {
		t.Error("GetNode returned false")
	}
	if retrieved.Name() != "node1" {
		t.Errorf("expected node name 'node1', got '%s'", retrieved.Name())
	}
}

func TestAddDuplicateNode(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node1")

	if err := g.AddNode(n1); err != nil {
		t.Errorf("AddNode failed: %v", err)
	}

	if err := g.AddNode(n2); err == nil {
		t.Error("expected error for duplicate node")
	}
}

func TestAddNodeEmptyName(t *testing.T) {
	g := New()
	n := node.NewFuncNode("", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	if err := g.AddNode(n); err == nil {
		t.Error("expected error for empty node name")
	}
}

func TestRemoveNode(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddEdge("node1", "node2")

	if err := g.RemoveNode("node1"); err != nil {
		t.Errorf("RemoveNode failed: %v", err)
	}

	if g.NodeCount() != 1 {
		t.Errorf("expected 1 node, got %d", g.NodeCount())
	}

	if g.EdgeCount() != 0 {
		t.Errorf("expected 0 edges, got %d", g.EdgeCount())
	}

	// 确保node2的依赖被清除
	deps := g.GetNodeDependencies("node2")
	if len(deps) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(deps))
	}
}

func TestRemoveNonexistentNode(t *testing.T) {
	g := New()

	if err := g.RemoveNode("nonexistent"); err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestAddEdge(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")

	g.AddNode(n1)
	g.AddNode(n2)

	if err := g.AddEdge("node1", "node2"); err != nil {
		t.Errorf("AddEdge failed: %v", err)
	}

	if g.EdgeCount() != 1 {
		t.Errorf("expected 1 edge, got %d", g.EdgeCount())
	}

	// 测试依赖关系
	deps := g.GetNodeDependencies("node2")
	if len(deps) != 1 || deps[0] != "node1" {
		t.Errorf("expected node2 to depend on node1")
	}

	dependents := g.GetNodeDependents("node1")
	if len(dependents) != 1 || dependents[0] != "node2" {
		t.Errorf("expected node1 to have node2 as dependent")
	}
}

func TestAddEdgeNonexistentNode(t *testing.T) {
	g := New()

	if err := g.AddEdge("node1", "node2"); err == nil {
		t.Error("expected error for nonexistent nodes")
	}
}

func TestAddEdgeSelfLoop(t *testing.T) {
	g := New()
	n := createTestNode("node1")
	g.AddNode(n)

	if err := g.AddEdge("node1", "node1"); err == nil {
		t.Error("expected error for self-loop")
	}
}

func TestAddDuplicateEdge(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddEdge("node1", "node2")

	if err := g.AddEdge("node1", "node2"); err == nil {
		t.Error("expected error for duplicate edge")
	}
}

func TestCycleDetection(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")
	n3 := createTestNode("node3")

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddNode(n3)

	g.AddEdge("node1", "node2")
	g.AddEdge("node2", "node3")

	// 添加会形成环的边
	if err := g.AddEdge("node3", "node1"); err == nil {
		t.Error("expected error for cycle detection")
	}

	// 确保图仍然是DAG
	if g.HasCycle() {
		t.Error("graph should not have cycle")
	}
}

func TestGetNodes(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")

	g.AddNode(n1)
	g.AddNode(n2)

	nodes := g.GetNodes()
	if len(nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(nodes))
	}
}

func TestGetEdges(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")
	n3 := createTestNode("node3")

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddNode(n3)

	g.AddEdge("node1", "node2")
	g.AddEdge("node2", "node3")

	edges := g.GetEdges()
	if len(edges) != 2 {
		t.Errorf("expected 2 edges, got %d", len(edges))
	}
}

func TestTopologicalSort(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")
	n3 := createTestNode("node3")

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddNode(n3)

	g.AddEdge("node1", "node2")
	g.AddEdge("node2", "node3")

	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Errorf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Errorf("expected 3 nodes, got %d", len(sorted))
	}

	// 验证顺序：node1 -> node2 -> node3
	if sorted[0].Name() != "node1" {
		t.Errorf("expected first node to be 'node1', got '%s'", sorted[0].Name())
	}
	if sorted[1].Name() != "node2" {
		t.Errorf("expected second node to be 'node2', got '%s'", sorted[1].Name())
	}
	if sorted[2].Name() != "node3" {
		t.Errorf("expected third node to be 'node3', got '%s'", sorted[2].Name())
	}
}

func TestTopologicalSortWithCycle(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")

	g.AddNode(n1)
	g.AddNode(n2)

	// 手动创建环（绕过AddEdge的检查）
	g.(*dagGraph).edges["node1"]["node2"] = true
	g.(*dagGraph).edges["node2"]["node1"] = true
	g.(*dagGraph).reverse["node2"]["node1"] = true
	g.(*dagGraph).reverse["node1"]["node2"] = true
	g.(*dagGraph).edgeList = []Edge{
		{From: "node1", To: "node2"},
		{From: "node2", To: "node1"},
	}

	_, err := g.TopologicalSort()
	if err == nil {
		t.Error("expected error for cycle in topological sort")
	}
}

func TestFindRunnableNodes(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")
	n3 := createTestNode("node3")
	n4 := createTestNode("node4")

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddNode(n3)
	g.AddNode(n4)

	g.AddEdge("node1", "node2")
	g.AddEdge("node1", "node3")
	g.AddEdge("node2", "node4")
	g.AddEdge("node3", "node4")

	// 初始状态：只有node1可以运行
	completed := make(map[string]bool)
	runnable := g.FindRunnableNodes(completed)
	if len(runnable) != 1 || runnable[0].Name() != "node1" {
		t.Errorf("expected only node1 to be runnable, got %v", runnable)
	}

	// 完成node1后：node2和node3可以运行
	completed["node1"] = true
	runnable = g.FindRunnableNodes(completed)
	if len(runnable) != 2 {
		t.Errorf("expected 2 runnable nodes, got %d", len(runnable))
	}

	// 完成node2和node3后：node4可以运行
	completed["node2"] = true
	completed["node3"] = true
	runnable = g.FindRunnableNodes(completed)
	if len(runnable) != 1 || runnable[0].Name() != "node4" {
		t.Errorf("expected only node4 to be runnable, got %v", runnable)
	}
}

func TestClone(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddEdge("node1", "node2")

	cloned := g.Clone()

	// 验证克隆的节点
	if cloned.NodeCount() != 2 {
		t.Errorf("expected 2 nodes in clone, got %d", cloned.NodeCount())
	}

	// 验证克隆的边
	if cloned.EdgeCount() != 1 {
		t.Errorf("expected 1 edge in clone, got %d", cloned.EdgeCount())
	}

	// 修改原始不应影响克隆
	n3 := createTestNode("node3")
	g.AddNode(n3)

	if cloned.NodeCount() != 2 {
		t.Error("changes to original affected clone")
	}

	// 修改克隆不应影响原始
	n4 := createTestNode("node4")
	cloned.AddNode(n4)

	if g.NodeCount() != 3 {
		t.Error("changes to clone affected original")
	}
}

func TestValidate(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddEdge("node1", "node2")

	if err := g.Validate(); err != nil {
		t.Errorf("Validate failed: %v", err)
	}
}

func TestValidateWithCycle(t *testing.T) {
	g := New()
	n1 := createTestNode("node1")
	n2 := createTestNode("node2")

	g.AddNode(n1)
	g.AddNode(n2)

	// 手动创建环
	g.(*dagGraph).edges["node1"]["node2"] = true
	g.(*dagGraph).edges["node2"]["node1"] = true
	g.(*dagGraph).reverse["node2"]["node1"] = true
	g.(*dagGraph).reverse["node1"]["node2"] = true
	g.(*dagGraph).edgeList = []Edge{
		{From: "node1", To: "node2"},
		{From: "node2", To: "node1"},
	}

	if err := g.Validate(); err == nil {
		t.Error("expected error for cycle in validation")
	}
}

func TestNodeCount(t *testing.T) {
	g := New()

	if g.NodeCount() != 0 {
		t.Errorf("expected 0 nodes, got %d", g.NodeCount())
	}

	g.AddNode(createTestNode("node1"))
	g.AddNode(createTestNode("node2"))

	if g.NodeCount() != 2 {
		t.Errorf("expected 2 nodes, got %d", g.NodeCount())
	}
}

func TestEdgeCount(t *testing.T) {
	g := New()

	if g.EdgeCount() != 0 {
		t.Errorf("expected 0 edges, got %d", g.EdgeCount())
	}

	g.AddNode(createTestNode("node1"))
	g.AddNode(createTestNode("node2"))
	g.AddEdge("node1", "node2")

	if g.EdgeCount() != 1 {
		t.Errorf("expected 1 edge, got %d", g.EdgeCount())
	}
}

func TestGraphInterface(t *testing.T) {
	// 确保 dagGraph 实现了 Graph 接口
	var _ Graph = New()
}

// 基准测试
func BenchmarkAddNode(b *testing.B) {
	g := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		n := createTestNode(fmt.Sprintf("node%d", i))
		g.AddNode(n)
	}
}

func BenchmarkAddEdge(b *testing.B) {
	g := New()
	for i := 0; i < 1000; i++ {
		g.AddNode(createTestNode(fmt.Sprintf("node%d", i)))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		from := fmt.Sprintf("node%d", i%1000)
		to := fmt.Sprintf("node%d", (i+1)%1000)
		g.AddEdge(from, to)
	}
}

func BenchmarkTopologicalSort(b *testing.B) {
	g := New()
	for i := 0; i < 100; i++ {
		g.AddNode(createTestNode(fmt.Sprintf("node%d", i)))
	}
	for i := 0; i < 99; i++ {
		g.AddEdge(fmt.Sprintf("node%d", i), fmt.Sprintf("node%d", i+1))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g.TopologicalSort()
	}
}

func BenchmarkFindRunnableNodes(b *testing.B) {
	g := New()
	for i := 0; i < 100; i++ {
		g.AddNode(createTestNode(fmt.Sprintf("node%d", i)))
	}
	for i := 0; i < 99; i++ {
		g.AddEdge(fmt.Sprintf("node%d", i), fmt.Sprintf("node%d", i+1))
	}

	completed := make(map[string]bool)
	for i := 0; i < 50; i++ {
		completed[fmt.Sprintf("node%d", i)] = true
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g.FindRunnableNodes(completed)
	}
}

func BenchmarkClone(b *testing.B) {
	g := New()
	for i := 0; i < 100; i++ {
		g.AddNode(createTestNode(fmt.Sprintf("node%d", i)))
	}
	for i := 0; i < 99; i++ {
		g.AddEdge(fmt.Sprintf("node%d", i), fmt.Sprintf("node%d", i+1))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		g.Clone()
	}
}
