package rollback

import (
	"errors"
	"testing"

	"github.com/elstworkflow/context"
	"github.com/elstworkflow/core"
)

func TestNewRollbackManager(t *testing.T) {
	manager := NewRollbackManager()
	if manager == nil {
		t.Fatal("NewRollbackManager returned nil")
	}
	if manager.Size() != 0 {
		t.Errorf("expected size 0, got %d", manager.Size())
	}
}

func TestPushPop(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	// 测试Push
	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			return nil
		},
		Context: ctx,
	})

	if manager.Size() != 1 {
		t.Errorf("expected size 1, got %d", manager.Size())
	}

	// 测试Pop
	compensation, ok := manager.Pop()
	if !ok {
		t.Error("expected Pop to return true")
	}
	if compensation.NodeName != "node1" {
		t.Errorf("expected node name 'node1', got '%s'", compensation.NodeName)
	}
	if manager.Size() != 0 {
		t.Errorf("expected size 0, got %d", manager.Size())
	}
}

func TestPopEmpty(t *testing.T) {
	manager := NewRollbackManager()

	_, ok := manager.Pop()
	if ok {
		t.Error("expected Pop to return false for empty stack")
	}
}

func TestRollback(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	rolledBack := make([]string, 0)

	// 添加补偿操作
	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			rolledBack = append(rolledBack, "node1")
			return nil
		},
		Context: ctx,
	})

	manager.Push(Compensation{
		NodeName: "node2",
		RollbackFunc: func(ctx core.Context) error {
			rolledBack = append(rolledBack, "node2")
			return nil
		},
		Context: ctx,
	})

	// 执行回滚
	err := manager.Rollback(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// 验证回滚顺序（后进先出）
	if len(rolledBack) != 2 {
		t.Errorf("expected 2 rollbacks, got %d", len(rolledBack))
	}
	if rolledBack[0] != "node2" {
		t.Errorf("expected first rollback to be 'node2', got '%s'", rolledBack[0])
	}
	if rolledBack[1] != "node1" {
		t.Errorf("expected second rollback to be 'node1', got '%s'", rolledBack[1])
	}
}

func TestRollbackError(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	expectedErr := errors.New("rollback failed")

	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			return expectedErr
		},
		Context: ctx,
	})

	err := manager.Rollback(ctx)
	if err == nil {
		t.Error("expected error")
	}

	var rollbackErr *RollbackError
	if !errors.As(err, &rollbackErr) {
		t.Errorf("expected RollbackError, got %T", err)
	}
	if rollbackErr.NodeName != "node1" {
		t.Errorf("expected node name 'node1', got '%s'", rollbackErr.NodeName)
	}
}

func TestRollbackPartialError(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	// 第一个节点回滚失败
	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			return errors.New("node1 rollback failed")
		},
		Context: ctx,
	})

	// 第二个节点回滚成功
	manager.Push(Compensation{
		NodeName: "node2",
		RollbackFunc: func(ctx core.Context) error {
			return nil
		},
		Context: ctx,
	})

	err := manager.Rollback(ctx)
	if err == nil {
		t.Error("expected error")
	}

	var rollbackErr *RollbackError
	if !errors.As(err, &rollbackErr) {
		t.Errorf("expected RollbackError, got %T", err)
	}
	if !rollbackErr.Partial {
		t.Error("expected partial rollback")
	}
}

func TestRollbackUntil(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	rolledBack := make([]string, 0)

	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			rolledBack = append(rolledBack, "node1")
			return nil
		},
		Context: ctx,
	})

	manager.Push(Compensation{
		NodeName: "node2",
		RollbackFunc: func(ctx core.Context) error {
			rolledBack = append(rolledBack, "node2")
			return nil
		},
		Context: ctx,
	})

	manager.Push(Compensation{
		NodeName: "node3",
		RollbackFunc: func(ctx core.Context) error {
			rolledBack = append(rolledBack, "node3")
			return nil
		},
		Context: ctx,
	})

	// 回滚到node2（包括node2）
	err := manager.RollbackUntil("node2", ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(rolledBack) != 2 {
		t.Errorf("expected 2 rollbacks, got %d", len(rolledBack))
	}
	if rolledBack[0] != "node3" {
		t.Errorf("expected first rollback to be 'node3', got '%s'", rolledBack[0])
	}
	if rolledBack[1] != "node2" {
		t.Errorf("expected second rollback to be 'node2', got '%s'", rolledBack[1])
	}
}

func TestRollbackUntilNotFound(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			return nil
		},
		Context: ctx,
	})

	err := manager.RollbackUntil("nonexistent", ctx)
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestRollbackOnly(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	rolledBack := false

	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			rolledBack = true
			return nil
		},
		Context: ctx,
	})

	err := manager.RollbackOnly("node1", ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !rolledBack {
		t.Error("expected rollback to be called")
	}
}

func TestRollbackOnlyNotFound(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	err := manager.RollbackOnly("nonexistent", ctx)
	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestSkipRollback(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	rolledBack := false

	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			rolledBack = true
			return nil
		},
		Context: ctx,
	})

	// 跳过node1的回滚
	manager.SkipRollback("node1")

	err := manager.Rollback(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if rolledBack {
		t.Error("expected rollback to be skipped")
	}
}

func TestSkipRollbackInCompensation(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	rolledBack := false

	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			rolledBack = true
			return nil
		},
		SkipRollback: true,
		Context:      ctx,
	})

	err := manager.Rollback(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if rolledBack {
		t.Error("expected rollback to be skipped")
	}
}

func TestClear(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			return nil
		},
		Context: ctx,
	})

	if manager.Size() != 1 {
		t.Errorf("expected size 1, got %d", manager.Size())
	}

	manager.Clear()

	if manager.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", manager.Size())
	}
}

func TestHasCompensation(t *testing.T) {
	manager := NewRollbackManager()
	ctx := context.New()

	if manager.HasCompensation("node1") {
		t.Error("expected HasCompensation to return false")
	}

	manager.Push(Compensation{
		NodeName: "node1",
		RollbackFunc: func(ctx core.Context) error {
			return nil
		},
		Context: ctx,
	})

	if !manager.HasCompensation("node1") {
		t.Error("expected HasCompensation to return true")
	}
}

func TestRollbackErrorInterface(t *testing.T) {
	err := &RollbackError{
		NodeName: "test",
		Err:      errors.New("test error"),
		Partial:  true,
	}

	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}

	if err.Unwrap() == nil {
		t.Error("expected Unwrap to return error")
	}
}

func TestRollbackManagerInterface(t *testing.T) {
	// 确保 rollbackManager 实现了 RollbackManager 接口
	var _ RollbackManager = NewRollbackManager()
}

// 基准测试
func BenchmarkPush(b *testing.B) {
	manager := NewRollbackManager()
	ctx := context.New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.Push(Compensation{
			NodeName: "node",
			RollbackFunc: func(ctx core.Context) error {
				return nil
			},
			Context: ctx,
		})
	}
}

func BenchmarkPop(b *testing.B) {
	manager := NewRollbackManager()
	ctx := context.New()

	for i := 0; i < b.N; i++ {
		manager.Push(Compensation{
			NodeName: "node",
			RollbackFunc: func(ctx core.Context) error {
				return nil
			},
			Context: ctx,
		})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.Pop()
	}
}

func BenchmarkRollback(b *testing.B) {
	manager := NewRollbackManager()
	ctx := context.New()

	for i := 0; i < 100; i++ {
		manager.Push(Compensation{
			NodeName: "node",
			RollbackFunc: func(ctx core.Context) error {
				return nil
			},
			Context: ctx,
		})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.Rollback(ctx)
	}
}
