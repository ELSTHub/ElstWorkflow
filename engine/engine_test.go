package engine

import (
	"testing"

	"github.com/elstworkflow/builder"
	"github.com/elstworkflow/context"
	"github.com/elstworkflow/core"
)

func createTestWorkflow(t *testing.T) *builder.Workflow {
	t.Helper()

	wf, err := builder.New("test-workflow").
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return "result1", nil
		}).
		Node("node2", func(ctx core.Context) (interface{}, error) {
			return "result2", nil
		}).
		DependsOn("node2", "node1").
		Build()

	if err != nil {
		t.Fatalf("failed to build workflow: %v", err)
	}

	return wf
}

func TestNewEngine(t *testing.T) {
	e := New(nil)
	if e == nil {
		t.Fatal("New() returned nil")
	}
	if e.Status() != StatusIdle {
		t.Errorf("expected status Idle, got %v", e.Status())
	}
}

func TestEngineLoad(t *testing.T) {
	e := New(nil)
	wf := createTestWorkflow(t)

	if err := e.Load(wf); err != nil {
		t.Errorf("Load failed: %v", err)
	}
}

func TestEngineLoadNil(t *testing.T) {
	e := New(nil)

	// 运行未加载的工作流应该失败
	ctx := context.New()
	_, err := e.Run(ctx)
	if err == nil {
		t.Error("expected error for running without loading")
	}
}

func TestEngineRun(t *testing.T) {
	e := New(nil)
	wf := createTestWorkflow(t)

	if err := e.Load(wf); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	ctx := context.New()
	result, err := e.Run(ctx)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
	if result.Status != core.WorkflowCompleted {
		t.Errorf("expected status Completed, got %v", result.Status)
	}
	if len(result.NodeResults) != 2 {
		t.Errorf("expected 2 node results, got %d", len(result.NodeResults))
	}
}

func TestEngineRunTwice(t *testing.T) {
	e := New(nil)
	wf := createTestWorkflow(t)

	if err := e.Load(wf); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	ctx := context.New()
	_, err := e.Run(ctx)
	if err != nil {
		t.Errorf("first Run failed: %v", err)
	}

	// 第二次运行应该失败（需要重新加载）
	_, err = e.Run(ctx)
	if err == nil {
		t.Error("expected error for running twice without reloading")
	}
}

func TestEnginePause(t *testing.T) {
	e := New(nil)
	wf := createTestWorkflow(t)

	if err := e.Load(wf); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 空闲状态暂停应该失败
	if err := e.Pause(); err == nil {
		t.Error("expected error for pausing idle engine")
	}
}

func TestEngineCancel(t *testing.T) {
	e := New(nil)
	wf := createTestWorkflow(t)

	if err := e.Load(wf); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 空闲状态取消应该失败
	if err := e.Cancel(); err == nil {
		t.Error("expected error for canceling idle engine")
	}
}

func TestEngineStatus(t *testing.T) {
	e := New(nil)

	if e.Status() != StatusIdle {
		t.Errorf("expected status Idle, got %v", e.Status())
	}
}

func TestEngineNodeResults(t *testing.T) {
	e := New(nil)
	wf := createTestWorkflow(t)

	if err := e.Load(wf); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	ctx := context.New()
	result, err := e.Run(ctx)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	nodeResults := e.NodeResults()
	if len(nodeResults) != len(result.NodeResults) {
		t.Errorf("expected %d node results, got %d", len(result.NodeResults), len(nodeResults))
	}
}

func TestEngineWithConfig(t *testing.T) {
	config := &Config{
		MaxParallel:   2,
		Executor:      nil, // 使用默认执行器
		SchedulerType: DAGScheduler,
	}

	e := New(config)
	wf := createTestWorkflow(t)

	if err := e.Load(wf); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	ctx := context.New()
	result, err := e.Run(ctx)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
	if result.Status != core.WorkflowCompleted {
		t.Errorf("expected status Completed, got %v", result.Status)
	}
}

func TestEngineWithDAGScheduler(t *testing.T) {
	config := &Config{
		SchedulerType: DAGScheduler,
	}

	e := New(config)

	// 创建并行工作流
	wf, err := builder.New("parallel-workflow").
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return "result1", nil
		}).
		Node("node2", func(ctx core.Context) (interface{}, error) {
			return "result2", nil
		}).
		Node("node3", func(ctx core.Context) (interface{}, error) {
			return "result3", nil
		}).
		DependsOn("node3", "node1").
		DependsOn("node3", "node2").
		Build()

	if err != nil {
		t.Fatalf("failed to build workflow: %v", err)
	}

	if err := e.Load(wf); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	ctx := context.New()
	result, err := e.Run(ctx)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}
	if result.Status != core.WorkflowCompleted {
		t.Errorf("expected status Completed, got %v", result.Status)
	}
}

func TestEngineInterface(t *testing.T) {
	// 确保 engine 实现了 Engine 接口
	var _ Engine = New(nil)
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusIdle, "Idle"},
		{StatusRunning, "Running"},
		{StatusPaused, "Paused"},
		{StatusCompleted, "Completed"},
		{StatusFailed, "Failed"},
		{StatusCancelled, "Cancelled"},
	}

	for _, tt := range tests {
		if tt.status.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.status.String())
		}
	}
}

func TestRunSimple(t *testing.T) {
	wf := createTestWorkflow(t)
	ctx := context.New()

	result, err := RunSimple(wf, ctx)
	if err != nil {
		t.Errorf("RunSimple failed: %v", err)
	}
	if result.Status != core.WorkflowCompleted {
		t.Errorf("expected status Completed, got %v", result.Status)
	}
}

// 基准测试
func BenchmarkEngineRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		e := New(nil)

		wf, _ := builder.New("bench-workflow").
			Node("node1", func(ctx core.Context) (interface{}, error) {
				return nil, nil
			}).
			Node("node2", func(ctx core.Context) (interface{}, error) {
				return nil, nil
			}).
			DependsOn("node2", "node1").
			Build()

		e.Load(wf)
		e.Run(context.New())
	}
}
