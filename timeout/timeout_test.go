package timeout

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ELSTHub/elstworkflow/core"
)

func TestNewTimeoutManager(t *testing.T) {
	manager := NewTimeoutManager()
	if manager == nil {
		t.Fatal("NewTimeoutManager returned nil")
	}
}

func TestWithTimeout(t *testing.T) {
	manager := NewTimeoutManager()

	// 测试成功执行
	result, err := manager.WithTimeout(time.Second, func(ctx context.Context) (interface{}, error) {
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %v", result)
	}
}

func TestWithTimeoutTimeout(t *testing.T) {
	manager := NewTimeoutManager()

	// 测试超时
	_, err := manager.WithTimeout(10*time.Millisecond, func(ctx context.Context) (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	})

	if err == nil {
		t.Error("expected timeout error")
	}

	if !IsTimeoutError(err) {
		t.Errorf("expected TimeoutError, got %T", err)
	}
}

func TestWithTimeoutError(t *testing.T) {
	manager := NewTimeoutManager()

	expectedErr := errors.New("test error")
	_, err := manager.WithTimeout(time.Second, func(ctx context.Context) (interface{}, error) {
		return nil, expectedErr
	})

	if err == nil {
		t.Error("expected error")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestWithTimeoutZeroTimeout(t *testing.T) {
	manager := NewTimeoutManager()

	// 测试零超时（应该立即执行）
	result, err := manager.WithTimeout(0, func(ctx context.Context) (interface{}, error) {
		return "immediate", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "immediate" {
		t.Errorf("expected 'immediate', got %v", result)
	}
}

func TestWithTimeoutPolicy(t *testing.T) {
	manager := NewTimeoutManager()

	policy := &core.TimeoutPolicy{
		Timeout:   time.Second,
		OnTimeout: core.TimeoutCancel,
	}

	result, err := manager.WithTimeoutPolicy(policy, func(ctx context.Context) (interface{}, error) {
		return "with policy", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "with policy" {
		t.Errorf("expected 'with policy', got %v", result)
	}
}

func TestWithTimeoutPolicyNil(t *testing.T) {
	manager := NewTimeoutManager()

	result, err := manager.WithTimeoutPolicy(nil, func(ctx context.Context) (interface{}, error) {
		return "no policy", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "no policy" {
		t.Errorf("expected 'no policy', got %v", result)
	}
}

func TestCreateContext(t *testing.T) {
	manager := NewTimeoutManager()

	ctx, cancel := manager.CreateContext(time.Second)
	defer cancel()

	if ctx == nil {
		t.Fatal("expected context")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("expected deadline")
	}

	if time.Until(deadline) > time.Second+100*time.Millisecond {
		t.Errorf("deadline too far in the future")
	}
}

func TestCreateContextWithParent(t *testing.T) {
	manager := NewTimeoutManager()

	parent, parentCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer parentCancel()

	ctx, cancel := manager.CreateContextWithParent(parent, time.Second)
	defer cancel()

	if ctx == nil {
		t.Fatal("expected context")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("expected deadline")
	}

	// 子上下文的截止时间应该比父上下文早
	parentDeadline, _ := parent.Deadline()
	if deadline.After(parentDeadline) {
		t.Error("child deadline should be before parent deadline")
	}
}

func TestExecuteWithTimeout(t *testing.T) {
	result, err := ExecuteWithTimeout(time.Second, func(ctx context.Context) (interface{}, error) {
		return "quick", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "quick" {
		t.Errorf("expected 'quick', got %v", result)
	}
}

func TestExecuteWithTimeoutPolicy(t *testing.T) {
	policy := &core.TimeoutPolicy{
		Timeout: time.Second,
	}

	result, err := ExecuteWithTimeoutPolicy(policy, func(ctx context.Context) (interface{}, error) {
		return "policy", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "policy" {
		t.Errorf("expected 'policy', got %v", result)
	}
}

func TestIsTimeoutError(t *testing.T) {
	// 测试超时错误
	timeoutErr := &TimeoutError{
		Operation: "test",
		Timeout:   time.Second,
		Elapsed:   time.Second,
	}

	if !IsTimeoutError(timeoutErr) {
		t.Error("expected true for TimeoutError")
	}

	// 测试普通错误
	normalErr := errors.New("普通错误")
	if IsTimeoutError(normalErr) {
		t.Error("expected false for non-TimeoutError")
	}

	// 测试nil错误
	if IsTimeoutError(nil) {
		t.Error("expected false for nil")
	}
}

func TestWrapWithTimeout(t *testing.T) {
	fn := func() (interface{}, error) {
		return "wrapped", nil
	}

	wrapped := WrapWithTimeout(fn, time.Second)
	result, err := wrapped()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "wrapped" {
		t.Errorf("expected 'wrapped', got %v", result)
	}
}

func TestWrapWithTimeoutTimeout(t *testing.T) {
	fn := func() (interface{}, error) {
		time.Sleep(100 * time.Millisecond)
		return nil, nil
	}

	wrapped := WrapWithTimeout(fn, 10*time.Millisecond)
	_, err := wrapped()

	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestWrapWithTimeoutPolicy(t *testing.T) {
	fn := func() (interface{}, error) {
		return "policy wrapped", nil
	}

	policy := &core.TimeoutPolicy{
		Timeout: time.Second,
	}

	wrapped := WrapWithTimeoutPolicy(fn, policy)
	result, err := wrapped()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "policy wrapped" {
		t.Errorf("expected 'policy wrapped', got %v", result)
	}
}

func TestWrapWithContext(t *testing.T) {
	fn := func(ctx context.Context) (interface{}, error) {
		return "context wrapped", nil
	}

	wrapped := WrapWithContext(fn, time.Second)
	result, err := wrapped()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "context wrapped" {
		t.Errorf("expected 'context wrapped', got %v", result)
	}
}

func TestTimeoutError(t *testing.T) {
	err := &TimeoutError{
		Operation: "test-op",
		Timeout:   5 * time.Second,
		Elapsed:   5 * time.Second,
	}

	expected := "operation test-op timed out after 5s (elapsed: 5s)"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestTimeoutManagerInterface(t *testing.T) {
	// 确保 timeoutManager 实现了 TimeoutManager 接口
	var _ TimeoutManager = NewTimeoutManager()
}

// 基准测试
func BenchmarkWithTimeout(b *testing.B) {
	manager := NewTimeoutManager()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.WithTimeout(time.Second, func(ctx context.Context) (interface{}, error) {
			return nil, nil
		})
	}
}

func BenchmarkWithTimeoutPolicy(b *testing.B) {
	manager := NewTimeoutManager()
	policy := &core.TimeoutPolicy{
		Timeout: time.Second,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.WithTimeoutPolicy(policy, func(ctx context.Context) (interface{}, error) {
			return nil, nil
		})
	}
}

func BenchmarkExecuteWithTimeout(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ExecuteWithTimeout(time.Second, func(ctx context.Context) (interface{}, error) {
			return nil, nil
		})
	}
}

func BenchmarkWrapWithTimeout(b *testing.B) {
	fn := func() (interface{}, error) {
		return nil, nil
	}
	wrapped := WrapWithTimeout(fn, time.Second)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wrapped()
	}
}
