package node

import (
	"errors"
	"testing"
	"time"

	"github.com/ELSTHub/elstworkflow/context"
	"github.com/ELSTHub/elstworkflow/core"
)

func TestNewFuncNode(t *testing.T) {
	executed := false
	executeFn := func(ctx core.Context) (interface{}, error) {
		executed = true
		return "result", nil
	}

	node := NewFuncNode("test-node", executeFn)

	if node.Name() != "test-node" {
		t.Errorf("expected name 'test-node', got '%s'", node.Name())
	}

	if err := node.Validate(); err != nil {
		t.Errorf("validation failed: %v", err)
	}

	ctx := context.New()
	result, err := node.Execute(ctx)
	if err != nil {
		t.Errorf("execution failed: %v", err)
	}
	if result != "result" {
		t.Errorf("expected 'result', got %v", result)
	}
	if !executed {
		t.Error("execute function was not called")
	}
}

func TestNewFuncNodeWithRollback(t *testing.T) {
	rolledBack := false
	executeFn := func(ctx core.Context) (interface{}, error) {
		return "result", nil
	}
	rollbackFn := func(ctx core.Context) error {
		rolledBack = true
		return nil
	}

	node := NewFuncNodeWithRollback("test-node", executeFn, rollbackFn)

	ctx := context.New()
	_, err := node.Execute(ctx)
	if err != nil {
		t.Errorf("execution failed: %v", err)
	}

	err = node.Rollback(ctx)
	if err != nil {
		t.Errorf("rollback failed: %v", err)
	}
	if !rolledBack {
		t.Error("rollback function was not called")
	}
}

func TestFuncNodeNilExecute(t *testing.T) {
	node := NewFuncNode("test-node", nil)

	if err := node.Validate(); err == nil {
		t.Error("expected validation error for nil execute function")
	}
}

func TestStaticNode(t *testing.T) {
	node := NewStaticNode("static", 42)

	if node.Name() != "static" {
		t.Errorf("expected name 'static', got '%s'", node.Name())
	}

	ctx := context.New()
	result, err := node.Execute(ctx)
	if err != nil {
		t.Errorf("execution failed: %v", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %v", result)
	}

	// 测试回滚
	err = node.Rollback(ctx)
	if err != nil {
		t.Errorf("rollback failed: %v", err)
	}
}

func TestErrorNode(t *testing.T) {
	expectedErr := errors.New("test error")
	node := NewErrorNode("error-node", expectedErr)

	ctx := context.New()
	_, err := node.Execute(ctx)
	if err == nil {
		t.Error("expected error")
	}
	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestDelayNode(t *testing.T) {
	start := time.Now()
	delay := 10 * time.Millisecond

	executeFn := func(ctx core.Context) (interface{}, error) {
		return "delayed", nil
	}

	node := NewDelayNode("delay-node", delay, executeFn)

	ctx := context.New()
	result, err := node.Execute(ctx)
	if err != nil {
		t.Errorf("execution failed: %v", err)
	}
	if result != "delayed" {
		t.Errorf("expected 'delayed', got %v", result)
	}

	elapsed := time.Since(start)
	if elapsed < delay {
		t.Errorf("expected delay of at least %v, got %v", delay, elapsed)
	}
}

func TestNodeOptions(t *testing.T) {
	retryPolicy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   time.Second,
	}

	timeoutPolicy := &core.TimeoutPolicy{
		Timeout:   30 * time.Second,
		OnTimeout: core.TimeoutCancel,
	}

	metadata := core.Metadata{
		"key1": "value1",
		"key2": "value2",
	}

	node := NewFuncNode("test-node", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	},
		WithDescription("test description"),
		WithRetryPolicy(retryPolicy),
		WithTimeoutPolicy(timeoutPolicy),
		WithParallel(true),
		WithMetadata(metadata),
		WithDependencies("dep1", "dep2"),
		WithCondition(func(ctx core.Context) (bool, error) {
			return true, nil
		}),
	)

	opts := node.Options()

	if opts.Name != "test-node" {
		t.Errorf("expected name 'test-node', got '%s'", opts.Name)
	}
	if opts.Description != "test description" {
		t.Errorf("expected description 'test description', got '%s'", opts.Description)
	}
	if opts.RetryPolicy != retryPolicy {
		t.Error("retry policy mismatch")
	}
	if opts.TimeoutPolicy != timeoutPolicy {
		t.Error("timeout policy mismatch")
	}
	if !opts.Parallel {
		t.Error("expected parallel to be true")
	}
	if opts.Metadata["key1"] != "value1" {
		t.Error("metadata key1 mismatch")
	}
	if len(opts.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(opts.Dependencies))
	}
	if opts.Condition == nil {
		t.Error("expected condition to be set")
	}
}

func TestNodeValidation(t *testing.T) {
	// 测试空名称
	node := NewFuncNode("", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	if err := node.Validate(); err == nil {
		t.Error("expected validation error for empty name")
	}
}

func TestNodeInterface(t *testing.T) {
	// 确保所有节点类型都实现了Node接口
	var _ Node = NewFuncNode("test", nil)
	var _ Node = NewFuncNodeWithRollback("test", nil, nil)
	var _ Node = NewStaticNode("test", nil)
	var _ Node = NewErrorNode("test", nil)
	var _ Node = NewDelayNode("test", 0, nil)
}

// 基准测试
func BenchmarkFuncNodeExecute(b *testing.B) {
	node := NewFuncNode("bench", func(ctx core.Context) (interface{}, error) {
		return "result", nil
	})

	ctx := context.New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		node.Execute(ctx)
	}
}

func BenchmarkStaticNodeExecute(b *testing.B) {
	node := NewStaticNode("bench", 42)
	ctx := context.New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		node.Execute(ctx)
	}
}

func BenchmarkNodeWithOptions(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFuncNode("bench",
			func(ctx core.Context) (interface{}, error) { return nil, nil },
			WithDescription("benchmark"),
			WithParallel(true),
			WithMetadata(core.Metadata{"key": "value"}),
		)
	}
}
