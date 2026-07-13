package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/elstworkflow/core"
)

func TestNewFixedRetryer(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   time.Second,
	}

	retryer := NewFixedRetryer(policy)
	if retryer == nil {
		t.Fatal("NewFixedRetryer returned nil")
	}

	if retryer.MaxAttempts() != 3 {
		t.Errorf("expected max attempts 3, got %d", retryer.MaxAttempts())
	}

	// 测试固定间隔
	for i := 0; i < 5; i++ {
		interval := retryer.NextInterval(i)
		if interval != time.Second {
			t.Errorf("expected interval 1s, got %v", interval)
		}
	}
}

func TestFixedRetryerShouldRetry(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   time.Second,
	}

	retryer := NewFixedRetryer(policy)
	err := errors.New("test error")

	// 应该重试
	if !retryer.ShouldRetry(err, 0) {
		t.Error("expected ShouldRetry to return true for attempt 0")
	}
	if !retryer.ShouldRetry(err, 1) {
		t.Error("expected ShouldRetry to return true for attempt 1")
	}
	if !retryer.ShouldRetry(err, 2) {
		t.Error("expected ShouldRetry to return true for attempt 2")
	}

	// 超过最大重试次数
	if retryer.ShouldRetry(err, 3) {
		t.Error("expected ShouldRetry to return false for attempt 3")
	}
}

func TestFixedRetryerWithRetryable(t *testing.T) {
	retryableErr := errors.New("retryable error")
	nonRetryableErr := errors.New("non-retryable error")

	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   time.Second,
		Retryable: func(err error) bool {
			return err.Error() == "retryable error"
		},
	}

	retryer := NewFixedRetryer(policy)

	if !retryer.ShouldRetry(retryableErr, 0) {
		t.Error("expected ShouldRetry to return true for retryable error")
	}
	if retryer.ShouldRetry(nonRetryableErr, 0) {
		t.Error("expected ShouldRetry to return false for non-retryable error")
	}
}

func TestNewExponentialRetryer(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:    core.RetryExponential,
		MaxRetries:  3,
		Interval:    time.Second,
		Multiplier:  2.0,
		MaxInterval: 10 * time.Second,
	}

	retryer := NewExponentialRetryer(policy)
	if retryer == nil {
		t.Fatal("NewExponentialRetryer returned nil")
	}

	// 测试指数退避间隔
	expectedIntervals := []time.Duration{
		time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
	}

	for i, expected := range expectedIntervals {
		interval := retryer.NextInterval(i)
		if interval != expected {
			t.Errorf("attempt %d: expected interval %v, got %v", i, expected, interval)
		}
	}
}

func TestExponentialRetryerMaxInterval(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:    core.RetryExponential,
		MaxRetries:  5,
		Interval:    time.Second,
		Multiplier:  2.0,
		MaxInterval: 5 * time.Second,
	}

	retryer := NewExponentialRetryer(policy)

	// 测试最大间隔限制
	for i := 0; i < 10; i++ {
		interval := retryer.NextInterval(i)
		if interval > 5*time.Second {
			t.Errorf("attempt %d: interval %v exceeds max interval", i, interval)
		}
	}
}

func TestNewCustomRetryer(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryCustom,
		MaxRetries: 3,
		Interval:   time.Second,
	}

	intervalFunc := func(attempt int) time.Duration {
		return time.Duration(attempt+1) * 500 * time.Millisecond
	}

	retryer := NewCustomRetryer(policy, intervalFunc)
	if retryer == nil {
		t.Fatal("NewCustomRetryer returned nil")
	}

	// 测试自定义间隔
	expectedIntervals := []time.Duration{
		500 * time.Millisecond,
		1000 * time.Millisecond,
		1500 * time.Millisecond,
	}

	for i, expected := range expectedIntervals {
		interval := retryer.NextInterval(i)
		if interval != expected {
			t.Errorf("attempt %d: expected interval %v, got %v", i, expected, interval)
		}
	}
}

func TestNewRetryer(t *testing.T) {
	// 测试固定策略
	fixedPolicy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   time.Second,
	}
	retryer := NewRetryer(fixedPolicy)
	if _, ok := retryer.(*fixedRetryer); !ok {
		t.Error("expected fixedRetryer")
	}

	// 测试指数策略
	exponentialPolicy := &core.RetryPolicy{
		Strategy:   core.RetryExponential,
		MaxRetries: 3,
		Interval:   time.Second,
		Multiplier: 2.0,
	}
	retryer = NewRetryer(exponentialPolicy)
	if _, ok := retryer.(*exponentialRetryer); !ok {
		t.Error("expected exponentialRetryer")
	}

	// 测试自定义策略
	customPolicy := &core.RetryPolicy{
		Strategy:   core.RetryCustom,
		MaxRetries: 3,
		Interval:   time.Second,
	}
	retryer = NewRetryer(customPolicy)
	if _, ok := retryer.(*customRetryer); !ok {
		t.Error("expected customRetryer")
	}

	// 测试nil策略
	retryer = NewRetryer(nil)
	if retryer != nil {
		t.Error("expected nil for nil policy")
	}
}

func TestExecute(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   10 * time.Millisecond,
	}

	retryer := NewFixedRetryer(policy)

	// 测试成功执行
	attempts := 0
	result, err := Execute(retryer, func() (interface{}, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("not ready")
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %v", result)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestExecuteMaxRetriesExceeded(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 2,
		Interval:   10 * time.Millisecond,
	}

	retryer := NewFixedRetryer(policy)

	attempts := 0
	_, err := Execute(retryer, func() (interface{}, error) {
		attempts++
		return nil, errors.New("always fail")
	})

	if err == nil {
		t.Error("expected error")
	}
	if attempts != 3 { // 初始尝试 + 2次重试
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestExecuteWithNilRetryer(t *testing.T) {
	attempts := 0
	result, err := Execute(nil, func() (interface{}, error) {
		attempts++
		return "result", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "result" {
		t.Errorf("expected 'result', got %v", result)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestExecuteWithCallback(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 2,
		Interval:   10 * time.Millisecond,
	}

	retryer := NewFixedRetryer(policy)

	retryAttempts := 0
	attempts := 0

	result, err := ExecuteWithCallback(retryer,
		func() (interface{}, error) {
			attempts++
			if attempts < 3 {
				return nil, errors.New("not ready")
			}
			return "success", nil
		},
		func(attempt int, err error, nextInterval time.Duration) {
			retryAttempts++
		},
	)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %v", result)
	}
	if retryAttempts != 2 {
		t.Errorf("expected 2 retry callbacks, got %d", retryAttempts)
	}
}

func TestRetryerInterface(t *testing.T) {
	// 确保所有重试器类型都实现了Retryer接口
	var _ Retryer = NewFixedRetryer(&core.RetryPolicy{})
	var _ Retryer = NewExponentialRetryer(&core.RetryPolicy{})
	var _ Retryer = NewCustomRetryer(&core.RetryPolicy{}, nil)
}

func TestRetryerReset(t *testing.T) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   time.Second,
	}

	retryer := NewFixedRetryer(policy)
	retryer.Reset() // 应该不会panic

	expRetryer := NewExponentialRetryer(policy)
	expRetryer.Reset() // 应该不会panic
}

// 基准测试
func BenchmarkFixedRetryerNextInterval(b *testing.B) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   time.Second,
	}
	retryer := NewFixedRetryer(policy)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		retryer.NextInterval(i)
	}
}

func BenchmarkExponentialRetryerNextInterval(b *testing.B) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryExponential,
		MaxRetries: 3,
		Interval:   time.Second,
		Multiplier: 2.0,
	}
	retryer := NewExponentialRetryer(policy)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		retryer.NextInterval(i)
	}
}

func BenchmarkExecute(b *testing.B) {
	policy := &core.RetryPolicy{
		Strategy:   core.RetryFixed,
		MaxRetries: 3,
		Interval:   time.Millisecond,
	}
	retryer := NewFixedRetryer(policy)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Execute(retryer, func() (interface{}, error) {
			return "result", nil
		})
	}
}
