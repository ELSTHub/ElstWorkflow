// Package retry 提供了重试策略的实现。
// 支持固定间隔、指数退避和自定义重试策略。
package retry

import (
	"math"
	"time"

	"github.com/elstworkflow/core"
)

// Retryer 定义重试器接口
type Retryer interface {
	// NextInterval 计算下一次重试的间隔
	NextInterval(attempt int) time.Duration
	// ShouldRetry 判断是否应该重试
	ShouldRetry(err error, attempt int) bool
	// MaxAttempts 返回最大重试次数
	MaxAttempts() int
	// Reset 重置重试器状态
	Reset()
}

// fixedRetryer 固定间隔重试器
type fixedRetryer struct {
	policy *core.RetryPolicy
}

// NewFixedRetryer 创建固定间隔重试器
func NewFixedRetryer(policy *core.RetryPolicy) Retryer {
	return &fixedRetryer{
		policy: policy,
	}
}

// NextInterval 返回固定间隔
func (r *fixedRetryer) NextInterval(attempt int) time.Duration {
	return r.policy.Interval
}

// ShouldRetry 判断是否应该重试
func (r *fixedRetryer) ShouldRetry(err error, attempt int) bool {
	if attempt >= r.policy.MaxRetries {
		return false
	}
	if r.policy.Retryable != nil {
		return r.policy.Retryable(err)
	}
	return true
}

// MaxAttempts 返回最大重试次数
func (r *fixedRetryer) MaxAttempts() int {
	return r.policy.MaxRetries
}

// Reset 重置重试器状态
func (r *fixedRetryer) Reset() {
	// 固定间隔重试器无需重置状态
}

// exponentialRetryer 指数退避重试器
type exponentialRetryer struct {
	policy          *core.RetryPolicy
	currentInterval time.Duration
}

// NewExponentialRetryer 创建指数退避重试器
func NewExponentialRetryer(policy *core.RetryPolicy) Retryer {
	return &exponentialRetryer{
		policy:          policy,
		currentInterval: policy.Interval,
	}
}

// NextInterval 返回指数退避间隔
func (r *exponentialRetryer) NextInterval(attempt int) time.Duration {
	if attempt == 0 {
		return r.policy.Interval
	}

	multiplier := r.policy.Multiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	interval := float64(r.policy.Interval) * math.Pow(multiplier, float64(attempt))

	// 应用最大间隔限制
	if r.policy.MaxInterval > 0 && time.Duration(interval) > r.policy.MaxInterval {
		return r.policy.MaxInterval
	}

	return time.Duration(interval)
}

// ShouldRetry 判断是否应该重试
func (r *exponentialRetryer) ShouldRetry(err error, attempt int) bool {
	if attempt >= r.policy.MaxRetries {
		return false
	}
	if r.policy.Retryable != nil {
		return r.policy.Retryable(err)
	}
	return true
}

// MaxAttempts 返回最大重试次数
func (r *exponentialRetryer) MaxAttempts() int {
	return r.policy.MaxRetries
}

// Reset 重置重试器状态
func (r *exponentialRetryer) Reset() {
	r.currentInterval = r.policy.Interval
}

// customRetryer 自定义重试器
type customRetryer struct {
	policy       *core.RetryPolicy
	intervalFunc func(attempt int) time.Duration
}

// NewCustomRetryer 创建自定义重试器
func NewCustomRetryer(policy *core.RetryPolicy, intervalFunc func(attempt int) time.Duration) Retryer {
	return &customRetryer{
		policy:       policy,
		intervalFunc: intervalFunc,
	}
}

// NextInterval 使用自定义函数计算间隔
func (r *customRetryer) NextInterval(attempt int) time.Duration {
	if r.intervalFunc != nil {
		return r.intervalFunc(attempt)
	}
	return r.policy.Interval
}

// ShouldRetry 判断是否应该重试
func (r *customRetryer) ShouldRetry(err error, attempt int) bool {
	if attempt >= r.policy.MaxRetries {
		return false
	}
	if r.policy.Retryable != nil {
		return r.policy.Retryable(err)
	}
	return true
}

// MaxAttempts 返回最大重试次数
func (r *customRetryer) MaxAttempts() int {
	return r.policy.MaxRetries
}

// Reset 重置重试器状态
func (r *customRetryer) Reset() {
	// 自定义重试器无需重置状态
}

// NewRetryer 根据策略创建重试器
func NewRetryer(policy *core.RetryPolicy) Retryer {
	if policy == nil {
		return nil
	}

	switch policy.Strategy {
	case core.RetryFixed:
		return NewFixedRetryer(policy)
	case core.RetryExponential:
		return NewExponentialRetryer(policy)
	case core.RetryCustom:
		return NewCustomRetryer(policy, nil)
	default:
		return NewFixedRetryer(policy)
	}
}

// Execute 执行带重试的操作
func Execute(retryer Retryer, fn func() (interface{}, error)) (interface{}, error) {
	if retryer == nil {
		return fn()
	}

	var lastErr error
	for attempt := 0; attempt <= retryer.MaxAttempts(); attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !retryer.ShouldRetry(err, attempt) {
			break
		}

		if attempt < retryer.MaxAttempts() {
			interval := retryer.NextInterval(attempt)
			time.Sleep(interval)
		}
	}

	return nil, lastErr
}

// ExecuteWithCallback 执行带重试和回调的操作
func ExecuteWithCallback(retryer Retryer, fn func() (interface{}, error), onRetry func(attempt int, err error, nextInterval time.Duration)) (interface{}, error) {
	if retryer == nil {
		return fn()
	}

	var lastErr error
	for attempt := 0; attempt <= retryer.MaxAttempts(); attempt++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if !retryer.ShouldRetry(err, attempt) {
			break
		}

		if attempt < retryer.MaxAttempts() {
			interval := retryer.NextInterval(attempt)
			if onRetry != nil {
				onRetry(attempt, err, interval)
			}
			time.Sleep(interval)
		}
	}

	return nil, lastErr
}
