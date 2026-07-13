package persistence

import (
	"testing"
	"time"

	"github.com/elstworkflow/core"
)

func TestNewMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	if store == nil {
		t.Fatal("NewMemoryStore() returned nil")
	}
}

func TestMemoryStoreWorkflow(t *testing.T) {
	store := NewMemoryStore()

	// 保存工作流
	data := []byte(`{"name":"test"}`)
	if err := store.SaveWorkflow("wf1", data); err != nil {
		t.Errorf("SaveWorkflow failed: %v", err)
	}

	// 加载工作流
	loaded, err := store.LoadWorkflow("wf1")
	if err != nil {
		t.Errorf("LoadWorkflow failed: %v", err)
	}
	if string(loaded) != string(data) {
		t.Errorf("expected '%s', got '%s'", string(data), string(loaded))
	}
}

func TestMemoryStoreWorkflowNotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.LoadWorkflow("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent workflow")
	}
}

func TestMemoryStoreDeleteWorkflow(t *testing.T) {
	store := NewMemoryStore()

	store.SaveWorkflow("wf1", []byte("data"))

	if err := store.DeleteWorkflow("wf1"); err != nil {
		t.Errorf("DeleteWorkflow failed: %v", err)
	}

	_, err := store.LoadWorkflow("wf1")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestMemoryStoreListWorkflows(t *testing.T) {
	store := NewMemoryStore()

	store.SaveWorkflow("wf1", []byte("data1"))
	store.SaveWorkflow("wf2", []byte("data2"))
	store.SaveWorkflow("wf3", []byte("data3"))

	ids, err := store.ListWorkflows()
	if err != nil {
		t.Errorf("ListWorkflows failed: %v", err)
	}
	if len(ids) != 3 {
		t.Errorf("expected 3 workflows, got %d", len(ids))
	}
}

func TestMemoryStoreNode(t *testing.T) {
	store := NewMemoryStore()

	// 保存节点状态
	data := []byte(`{"status":"completed"}`)
	if err := store.SaveNode("wf1", "node1", data); err != nil {
		t.Errorf("SaveNode failed: %v", err)
	}

	// 加载节点状态
	loaded, err := store.LoadNode("wf1", "node1")
	if err != nil {
		t.Errorf("LoadNode failed: %v", err)
	}
	if string(loaded) != string(data) {
		t.Errorf("expected '%s', got '%s'", string(data), string(loaded))
	}
}

func TestMemoryStoreNodeNotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.LoadNode("wf1", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestMemoryStoreDeleteNode(t *testing.T) {
	store := NewMemoryStore()

	store.SaveNode("wf1", "node1", []byte("data"))

	if err := store.DeleteNode("wf1", "node1"); err != nil {
		t.Errorf("DeleteNode failed: %v", err)
	}

	_, err := store.LoadNode("wf1", "node1")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestMemoryStoreListNodes(t *testing.T) {
	store := NewMemoryStore()

	store.SaveNode("wf1", "node1", []byte("data1"))
	store.SaveNode("wf1", "node2", []byte("data2"))

	names, err := store.ListNodes("wf1")
	if err != nil {
		t.Errorf("ListNodes failed: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(names))
	}
}

func TestMemoryStoreCheckpoint(t *testing.T) {
	store := NewMemoryStore()

	checkpoint := &Checkpoint{
		WorkflowID:     "wf1",
		Status:         core.WorkflowRunning,
		CompletedNodes: []string{"node1", "node2"},
		FailedNodes:    []string{},
		NodeResults:    make(map[string]*core.NodeResult),
		ContextData:    map[string]interface{}{"key": "value"},
		CreatedAt:      time.Now(),
	}

	// 保存检查点
	if err := store.SaveCheckpoint("wf1", checkpoint); err != nil {
		t.Errorf("SaveCheckpoint failed: %v", err)
	}

	// 加载检查点
	loaded, err := store.LoadCheckpoint("wf1")
	if err != nil {
		t.Errorf("LoadCheckpoint failed: %v", err)
	}
	if loaded.WorkflowID != "wf1" {
		t.Errorf("expected workflow ID 'wf1', got '%s'", loaded.WorkflowID)
	}
	if loaded.Status != core.WorkflowRunning {
		t.Errorf("expected status Running, got %v", loaded.Status)
	}
	if len(loaded.CompletedNodes) != 2 {
		t.Errorf("expected 2 completed nodes, got %d", len(loaded.CompletedNodes))
	}
}

func TestMemoryStoreCheckpointNotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.LoadCheckpoint("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent checkpoint")
	}
}

func TestMemoryStoreDeleteCheckpoint(t *testing.T) {
	store := NewMemoryStore()

	checkpoint := &Checkpoint{
		WorkflowID: "wf1",
		Status:     core.WorkflowRunning,
	}

	store.SaveCheckpoint("wf1", checkpoint)

	if err := store.DeleteCheckpoint("wf1"); err != nil {
		t.Errorf("DeleteCheckpoint failed: %v", err)
	}

	_, err := store.LoadCheckpoint("wf1")
	if err == nil {
		t.Error("expected error after deletion")
	}
}

func TestFileStore(t *testing.T) {
	store := NewFileStore("/tmp/test")

	// FileStore 是预留实现，所有操作应该返回错误
	if err := store.SaveWorkflow("wf1", []byte("data")); err == nil {
		t.Error("expected error for unimplemented method")
	}

	if _, err := store.LoadWorkflow("wf1"); err == nil {
		t.Error("expected error for unimplemented method")
	}
}

func TestStoreInterface(t *testing.T) {
	// 确保所有存储类型都实现了Store接口
	var _ Store = NewMemoryStore()
	var _ Store = NewFileStore("")
}

func TestCheckpoint(t *testing.T) {
	checkpoint := &Checkpoint{
		WorkflowID:     "wf1",
		Status:         core.WorkflowCompleted,
		CompletedNodes: []string{"node1", "node2", "node3"},
		FailedNodes:    []string{},
		NodeResults: map[string]*core.NodeResult{
			"node1": {Status: core.NodeCompleted, Output: "result1"},
			"node2": {Status: core.NodeCompleted, Output: "result2"},
		},
		ContextData: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if checkpoint.WorkflowID != "wf1" {
		t.Errorf("expected 'wf1', got '%s'", checkpoint.WorkflowID)
	}
	if checkpoint.Status != core.WorkflowCompleted {
		t.Errorf("expected Completed, got %v", checkpoint.Status)
	}
}

// 基准测试
func BenchmarkSaveWorkflow(b *testing.B) {
	store := NewMemoryStore()
	data := []byte(`{"name":"test"}`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.SaveWorkflow("wf1", data)
	}
}

func BenchmarkLoadWorkflow(b *testing.B) {
	store := NewMemoryStore()
	store.SaveWorkflow("wf1", []byte(`{"name":"test"}`))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.LoadWorkflow("wf1")
	}
}

func BenchmarkSaveNode(b *testing.B) {
	store := NewMemoryStore()
	data := []byte(`{"status":"completed"}`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.SaveNode("wf1", "node1", data)
	}
}

func BenchmarkSaveCheckpoint(b *testing.B) {
	store := NewMemoryStore()
	checkpoint := &Checkpoint{
		WorkflowID: "wf1",
		Status:     core.WorkflowRunning,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.SaveCheckpoint("wf1", checkpoint)
	}
}
