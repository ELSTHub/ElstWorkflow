// Package engine 提供了工作流引擎的实现。
// 引擎是整个工作流的运行时，负责加载工作流、调度、执行、回滚等功能。
package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/ELSTHub/elstworkflow/builder"
	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/executor"
	"github.com/ELSTHub/elstworkflow/graph"
	"github.com/ELSTHub/elstworkflow/rollback"
	"github.com/ELSTHub/elstworkflow/scheduler"
)

// Status 表示引擎状态
type Status int

const (
	// StatusIdle 空闲状态
	StatusIdle Status = iota
	// StatusRunning 运行中
	StatusRunning
	// StatusPaused 已暂停
	StatusPaused
	// StatusCompleted 已完成
	StatusCompleted
	// StatusFailed 失败
	StatusFailed
	// StatusCancelled 已取消
	StatusCancelled
)

// String 返回状态的字符串表示
func (s Status) String() string {
	switch s {
	case StatusIdle:
		return "Idle"
	case StatusRunning:
		return "Running"
	case StatusPaused:
		return "Paused"
	case StatusCompleted:
		return "Completed"
	case StatusFailed:
		return "Failed"
	case StatusCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

// Engine 定义工作流引擎接口
type Engine interface {
	// Load 加载工作流
	Load(wf *builder.Workflow) error
	// Run 运行工作流
	Run(ctx core.Context) (*core.WorkflowResult, error)
	// Pause 暂停工作流
	Pause() error
	// Resume 恢复工作流
	Resume(ctx core.Context) (*core.WorkflowResult, error)
	// Cancel 取消工作流
	Cancel() error
	// Status 返回引擎状态
	Status() Status
	// NodeResults 返回节点执行结果
	NodeResults() map[string]*core.NodeResult
}

// Config 定义引擎配置
type Config struct {
	// MaxParallel 最大并行数
	MaxParallel int
	// Executor 执行器
	Executor executor.Executor
	// SchedulerType 调度器类型
	SchedulerType SchedulerType
}

// SchedulerType 定义调度器类型
type SchedulerType int

const (
	// SerialScheduler 串行调度器
	SerialScheduler SchedulerType = iota
	// DAGScheduler DAG调度器
	DAGScheduler
	// PriorityScheduler 优先级调度器
	PriorityScheduler
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxParallel:   1,
		Executor:      executor.New(),
		SchedulerType: SerialScheduler,
	}
}

// engine 是 Engine 接口的默认实现
type engine struct {
	mu          sync.RWMutex
	config      *Config
	workflow    *builder.Workflow
	scheduler   scheduler.Scheduler
	executor    executor.Executor
	rollbackMgr rollback.RollbackManager
	ctx         core.Context
	status      Status
	nodeResults map[string]*core.NodeResult
	startTime   time.Time
	endTime     time.Duration
}

// New 创建新的工作流引擎
func New(config *Config) Engine {
	if config == nil {
		config = DefaultConfig()
	}

	// 如果未指定执行器，使用默认执行器
	exec := config.Executor
	if exec == nil {
		exec = executor.New()
	}

	return &engine{
		config:      config,
		executor:    exec,
		rollbackMgr: rollback.NewRollbackManager(),
		status:      StatusIdle,
		nodeResults: make(map[string]*core.NodeResult),
	}
}

// Load 加载工作流
func (e *engine) Load(wf *builder.Workflow) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status == StatusRunning {
		return fmt.Errorf("cannot load workflow while engine is running")
	}

	// 验证工作流图
	if err := wf.Graph.Validate(); err != nil {
		return fmt.Errorf("workflow graph validation failed: %v", err)
	}

	// 创建调度器
	var err error
	e.scheduler, err = e.createScheduler(wf.Graph)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %v", err)
	}

	e.workflow = wf
	e.status = StatusIdle
	e.nodeResults = make(map[string]*core.NodeResult)

	return nil
}

// createScheduler 创建调度器
func (e *engine) createScheduler(g graph.Graph) (scheduler.Scheduler, error) {
	switch e.config.SchedulerType {
	case SerialScheduler:
		return scheduler.NewSerialScheduler(g)
	case DAGScheduler:
		return scheduler.NewDAGScheduler(g), nil
	case PriorityScheduler:
		return scheduler.NewPriorityScheduler(g, nil), nil
	default:
		return scheduler.NewSerialScheduler(g)
	}
}

// Run 运行工作流
func (e *engine) Run(ctx core.Context) (*core.WorkflowResult, error) {
	e.mu.Lock()
	if e.status == StatusRunning {
		e.mu.Unlock()
		return nil, fmt.Errorf("workflow is already running")
	}
	if e.workflow == nil {
		e.mu.Unlock()
		return nil, fmt.Errorf("no workflow loaded")
	}
	if e.status == StatusCompleted || e.status == StatusFailed || e.status == StatusCancelled {
		e.mu.Unlock()
		return nil, fmt.Errorf("workflow has already finished, reload to run again")
	}
	e.status = StatusRunning
	e.ctx = ctx
	e.startTime = time.Now()
	e.mu.Unlock()

	// 执行工作流
	err := e.execute(ctx)

	e.mu.Lock()
	defer e.mu.Unlock()

	endTime := time.Now()

	// 构建结果
	result := &core.WorkflowResult{
		NodeResults: e.nodeResults,
		StartTime:   e.startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(e.startTime),
	}

	if err != nil {
		e.status = StatusFailed
		result.Status = core.WorkflowFailed
		result.Error = err
	} else {
		e.status = StatusCompleted
		result.Status = core.WorkflowCompleted
	}

	return result, nil
}

// execute 执行工作流
func (e *engine) execute(ctx core.Context) error {
	for {
		// 检查是否被取消
		e.mu.RLock()
		if e.status == StatusCancelled {
			e.mu.RUnlock()
			return fmt.Errorf("workflow cancelled")
		}
		if e.status == StatusPaused {
			e.mu.RUnlock()
			time.Sleep(10 * time.Millisecond)
			continue
		}
		e.mu.RUnlock()

		// 获取下一个可执行节点
		n, ok := e.scheduler.Next()
		if !ok {
			// 没有更多可执行节点
			break
		}

		// 执行节点
		result, err := e.executor.Execute(ctx, n)
		if err != nil {
			e.scheduler.MarkFailed(n.Name())
			e.mu.Lock()
			e.nodeResults[n.Name()] = &core.NodeResult{
				Status:    core.NodeFailed,
				Error:     err,
				StartTime: time.Now(),
				EndTime:   time.Now(),
			}
			e.mu.Unlock()
			return fmt.Errorf("node %s failed: %v", n.Name(), err)
		}

		// 标记节点完成
		e.scheduler.MarkCompleted(n.Name())
		e.mu.Lock()
		e.nodeResults[n.Name()] = result
		e.mu.Unlock()

		// 将输出存入上下文
		if result != nil && result.Output != nil {
			ctx.Put(n.Name(), result.Output)
		}
	}

	return nil
}

// Pause 暂停工作流
func (e *engine) Pause() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != StatusRunning {
		return fmt.Errorf("workflow is not running")
	}

	e.status = StatusPaused
	return nil
}

// Resume 恢复工作流
func (e *engine) Resume(ctx core.Context) (*core.WorkflowResult, error) {
	e.mu.Lock()
	if e.status != StatusPaused {
		e.mu.Unlock()
		return nil, fmt.Errorf("workflow is not paused")
	}
	e.status = StatusRunning
	e.mu.Unlock()

	// 继续执行
	err := e.execute(ctx)

	e.mu.Lock()
	defer e.mu.Unlock()

	endTime := time.Now()

	result := &core.WorkflowResult{
		NodeResults: e.nodeResults,
		StartTime:   e.startTime,
		EndTime:     endTime,
		Duration:    endTime.Sub(e.startTime),
	}

	if err != nil {
		e.status = StatusFailed
		result.Status = core.WorkflowFailed
		result.Error = err
	} else {
		e.status = StatusCompleted
		result.Status = core.WorkflowCompleted
	}

	return result, nil
}

// Cancel 取消工作流
func (e *engine) Cancel() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != StatusRunning && e.status != StatusPaused {
		return fmt.Errorf("workflow is not running or paused")
	}

	e.status = StatusCancelled
	return nil
}

// Status 返回引擎状态
func (e *engine) Status() Status {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.status
}

// NodeResults 返回节点执行结果
func (e *engine) NodeResults() map[string]*core.NodeResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	results := make(map[string]*core.NodeResult, len(e.nodeResults))
	for k, v := range e.nodeResults {
		results[k] = v
	}
	return results
}

// RunSimple 简单运行工作流的便捷函数
func RunSimple(wf *builder.Workflow, ctx core.Context) (*core.WorkflowResult, error) {
	e := New(nil)
	if err := e.Load(wf); err != nil {
		return nil, err
	}
	return e.Run(ctx)
}
