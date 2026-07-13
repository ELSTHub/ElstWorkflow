package executor

import (
	"errors"
	"testing"

	"github.com/ELSTHub/elstworkflow/context"
	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/node"
	"github.com/ELSTHub/elstworkflow/retry"
	"github.com/ELSTHub/elstworkflow/rollback"
)

func TestNewExecutor(t *testing.T) {
	e := New()
	if e == nil {
		t.Fatal("New() returned nil")
	}
}

func TestExecuteSuccess(t *testing.T) {
	e := New()
	ctx := context.New()

	n := node.NewFuncNode("test-node", func(ctx core.Context) (interface{}, error) {
		return "result", nil
	})

	result, err := e.Execute(ctx, n)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.Status != core.NodeCompleted {
		t.Errorf("expected status Completed, got %v", result.Status)
	}
	if result.Output != "result" {
		t.Errorf("expected output 'result', got %v", result.Output)
	}
}

func TestExecuteError(t *testing.T) {
	e := New()
	ctx := context.New()

	expectedErr := errors.New("test error")
	n := node.NewFuncNode("test-node", func(ctx core.Context) (interface{}, error) {
		return nil, expectedErr
	})

	_, err := e.Execute(ctx, n)
	if err == nil {
		t.Error("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestExecuteWithValidation(t *testing.T) {
	e := New()
	ctx := context.New()

	// 空名称的节点应该验证失败
	n := node.NewFuncNode("", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	_, err := e.Execute(ctx, n)
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestExecuteWithMiddleware(t *testing.T) {
	executed := false

	middleware := func(next ExecuteFunc) ExecuteFunc {
		return func(ctx core.Context, n node.Node) (*core.NodeResult, error) {
			executed = true
			return next(ctx, n)
		}
	}

	e := New(WithMiddleware(middleware))
	ctx := context.New()

	n := node.NewFuncNode("test-node", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	_, err := e.Execute(ctx, n)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !executed {
		t.Error("middleware was not executed")
	}
}

func TestExecuteWithRetry(t *testing.T) {
	attempts := 0

	retryPolicy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 2,
		Interval:   1, // 1ms for fast test
	}
	retryer := retry.NewFixedRetryer(retryPolicy)

	e := New(WithRetryer(retryer))
	ctx := context.New()

	n := node.NewFuncNode("test-node", func(ctx core.Context) (interface{}, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("not ready")
		}
		return "success", nil
	})

	result, err := e.Execute(ctx, n)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result.Output != "success" {
		t.Errorf("expected output 'success', got %v", result.Output)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestExecuteWithRollback(t *testing.T) {
	e := New()
	ctx := context.New()
	rollbackMgr := rollback.NewRollbackManager()

	rolledBack := false
	expectedErr := errors.New("execution failed")

	n := node.NewFuncNodeWithRollback("test-node",
		func(ctx core.Context) (interface{}, error) {
			return nil, expectedErr
		},
		func(ctx core.Context) error {
			rolledBack = true
			return nil
		},
	)

	_, err := e.ExecuteWithRollback(ctx, n, rollbackMgr)
	if err == nil {
		t.Error("expected error")
	}
	if !rolledBack {
		t.Error("expected rollback to be called")
	}
}

func TestExecutorInterface(t *testing.T) {
	// 确保 executor 实现了 Executor 接口
	var _ Executor = New()
}

// 基准测试
func BenchmarkExecute(b *testing.B) {
	e := New()
	ctx := context.New()

	n := node.NewFuncNode("bench", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Execute(ctx, n)
	}
}

func BenchmarkExecuteWithMiddleware(b *testing.B) {
	middleware := func(next ExecuteFunc) ExecuteFunc {
		return func(ctx core.Context, n node.Node) (*core.NodeResult, error) {
			return next(ctx, n)
		}
	}

	e := New(WithMiddleware(middleware))
	ctx := context.New()

	n := node.NewFuncNode("bench", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		e.Execute(ctx, n)
	}
}
