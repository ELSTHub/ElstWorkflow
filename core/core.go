// Package core 定义了工作流框架的核心类型和接口。
// 这些类型被所有其他包共享，是整个框架的基础。
package core

import (
	"time"
)

// WorkflowStatus 表示工作流的执行状态
type WorkflowStatus int

const (
	// WorkflowPending 工作流等待执行
	WorkflowPending WorkflowStatus = iota
	// WorkflowRunning 工作流正在运行
	WorkflowRunning
	// WorkflowCompleted 工作流已完成
	WorkflowCompleted
	// WorkflowFailed 工作流执行失败
	WorkflowFailed
	// WorkflowCancelled 工作流已取消
	WorkflowCancelled
	// WorkflowPaused 工作流已暂停
	WorkflowPaused
)

// String 返回工作流状态的字符串表示
func (s WorkflowStatus) String() string {
	switch s {
	case WorkflowPending:
		return "Pending"
	case WorkflowRunning:
		return "Running"
	case WorkflowCompleted:
		return "Completed"
	case WorkflowFailed:
		return "Failed"
	case WorkflowCancelled:
		return "Cancelled"
	case WorkflowPaused:
		return "Paused"
	default:
		return "Unknown"
	}
}

// NodeStatus 表示节点的执行状态
type NodeStatus int

const (
	// NodePending 节点等待执行
	NodePending NodeStatus = iota
	// NodeRunning 节点正在运行
	NodeRunning
	// NodeCompleted 节点已完成
	NodeCompleted
	// NodeFailed 节点执行失败
	NodeFailed
	// NodeSkipped 节点被跳过
	NodeSkipped
	// NodeCancelled 节点已取消
	NodeCancelled
	// NodeRollingBack 节点正在回滚
	NodeRollingBack
	// NodeRolledBack 节点已回滚
	NodeRolledBack
)

// String 返回节点状态的字符串表示
func (s NodeStatus) String() string {
	switch s {
	case NodePending:
		return "Pending"
	case NodeRunning:
		return "Running"
	case NodeCompleted:
		return "Completed"
	case NodeFailed:
		return "Failed"
	case NodeSkipped:
		return "Skipped"
	case NodeCancelled:
		return "Cancelled"
	case NodeRollingBack:
		return "RollingBack"
	case NodeRolledBack:
		return "RolledBack"
	default:
		return "Unknown"
	}
}

// RetryStrategy 定义重试策略类型
type RetryStrategy int

const (
	// RetryFixed 固定间隔重试
	RetryFixed RetryStrategy = iota
	// RetryExponential 指数退避重试
	RetryExponential
	// RetryCustom 自定义重试策略
	RetryCustom
)

// String 返回重试策略的字符串表示
func (s RetryStrategy) String() string {
	switch s {
	case RetryFixed:
		return "Fixed"
	case RetryExponential:
		return "Exponential"
	case RetryCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

// Metadata 是工作流和节点的元数据类型
type Metadata map[string]string

// Options 定义通用选项
type Options struct {
	// MaxParallel 最大并行数
	MaxParallel int
	// Timeout 默认超时时间
	Timeout time.Duration
	// RetryPolicy 默认重试策略
	RetryPolicy *RetryPolicy
	// Metadata 元数据
	Metadata Metadata
}

// RetryPolicy 定义重试策略
type RetryPolicy struct {
	// Strategy 重试策略类型
	Strategy RetryStrategy
	// MaxRetries 最大重试次数
	MaxRetries int
	// Interval 重试间隔
	Interval time.Duration
	// MaxInterval 最大重试间隔（用于指数退避）
	MaxInterval time.Duration
	// Multiplier 退避乘数（用于指数退避）
	Multiplier float64
	// Retryable 判断是否可重试的函数
	Retryable func(error) bool
}

// TimeoutPolicy 定义超时策略
type TimeoutPolicy struct {
	// Timeout 超时时间
	Timeout time.Duration
	// OnTimeout 超时时的处理方式
	OnTimeout TimeoutAction
}

// TimeoutAction 定义超时动作
type TimeoutAction int

const (
	// TimeoutCancel 超时后取消
	TimeoutCancel TimeoutAction = iota
	// TimeoutFail 超时后失败
	TimeoutFail
	// TimeoutRetry 超时后重试
	TimeoutRetry
)

// String 返回超时动作的字符串表示
func (a TimeoutAction) String() string {
	switch a {
	case TimeoutCancel:
		return "Cancel"
	case TimeoutFail:
		return "Fail"
	case TimeoutRetry:
		return "Retry"
	default:
		return "Unknown"
	}
}

// NodeResult 表示节点执行结果
type NodeResult struct {
	// Status 节点状态
	Status NodeStatus
	// Output 输出数据
	Output interface{}
	// Error 错误信息
	Error error
	// StartTime 开始时间
	StartTime time.Time
	// EndTime 结束时间
	EndTime time.Time
	// Duration 执行时长
	Duration time.Duration
	// RetryCount 重试次数
	RetryCount int
}

// WorkflowResult 表示工作流执行结果
type WorkflowResult struct {
	// Status 工作流状态
	Status WorkflowStatus
	// NodeResults 各节点执行结果
	NodeResults map[string]*NodeResult
	// Error 错误信息
	Error error
	// StartTime 开始时间
	StartTime time.Time
	// EndTime 结束时间
	EndTime time.Time
	// Duration 执行时长
	Duration time.Duration
}

// Context 定义工作流上下文接口
// 提供线程安全的键值存储，用于在节点间传递数据
type Context interface {
	// Put 设置键值对
	Put(key string, value interface{})
	// Get 获取键对应的值
	Get(key string) (interface{}, bool)
	// GetTyped 获取键对应的值并尝试转换为指定类型
	GetTyped(key string, target interface{}) (interface{}, error)
	// Delete 删除键值对
	Delete(key string)
	// Keys 返回所有键
	Keys() []string
	// Clone 克隆上下文
	Clone() Context
	// Has 检查键是否存在
	Has(key string) bool
	// Len 返回键值对数量
	Len() int
	// Clear 清空所有键值对
	Clear()
}

// Condition 定义条件函数类型，用于条件分支
type Condition func(ctx Context) (bool, error)

// NodeFunc 定义节点执行函数类型
type NodeFunc func(ctx Context) (interface{}, error)

// RollbackFunc 定义回滚函数类型
type RollbackFunc func(ctx Context) error
