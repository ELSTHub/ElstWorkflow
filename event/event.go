// Package event 提供了事件总线的实现。
// 支持工作流和节点生命周期事件的发布和订阅。
package event

import (
	"fmt"
	"sync"
	"time"
)

// EventType 定义事件类型
type EventType int

const (
	// WorkflowStarted 工作流开始事件
	WorkflowStarted EventType = iota
	// WorkflowFinished 工作流完成事件
	WorkflowFinished
	// WorkflowFailed 工作流失败事件
	WorkflowFailed
	// WorkflowCancelled 工作流取消事件
	WorkflowCancelled
	// NodeStarted 节点开始事件
	NodeStarted
	// NodeFinished 节点完成事件
	NodeFinished
	// NodeFailed 节点失败事件
	NodeFailed
	// NodeRetry 节点重试事件
	NodeRetry
	// NodeRollback 节点回滚事件
	NodeRollback
)

// String 返回事件类型的字符串表示
func (t EventType) String() string {
	switch t {
	case WorkflowStarted:
		return "WorkflowStarted"
	case WorkflowFinished:
		return "WorkflowFinished"
	case WorkflowFailed:
		return "WorkflowFailed"
	case WorkflowCancelled:
		return "WorkflowCancelled"
	case NodeStarted:
		return "NodeStarted"
	case NodeFinished:
		return "NodeFinished"
	case NodeFailed:
		return "NodeFailed"
	case NodeRetry:
		return "NodeRetry"
	case NodeRollback:
		return "NodeRollback"
	default:
		return "Unknown"
	}
}

// Event 表示一个事件
type Event struct {
	// Type 事件类型
	Type EventType
	// WorkflowName 工作流名称
	WorkflowName string
	// NodeName 节点名称
	NodeName string
	// Timestamp 事件时间戳
	Timestamp time.Time
	// Data 事件数据
	Data interface{}
	// Error 错误信息
	Error error
}

// Handler 定义事件处理函数类型
type Handler func(event Event)

// EventBus 定义事件总线接口
type EventBus interface {
	// Subscribe 订阅事件
	Subscribe(eventType EventType, handler Handler) string
	// Unsubscribe 取消订阅
	Unsubscribe(subscriptionID string) error
	// Publish 发布事件
	Publish(event Event)
	// Close 关闭事件总线
	Close()
}

// subscription 表示一个订阅
type subscription struct {
	id      string
	handler Handler
}

// eventBus 是 EventBus 接口的默认实现
type eventBus struct {
	mu            sync.RWMutex
	subscriptions map[EventType][]subscription
	closed        bool
	nextID        int
}

// NewEventBus 创建新的事件总线
func NewEventBus() EventBus {
	return &eventBus{
		subscriptions: make(map[EventType][]subscription),
	}
}

// Subscribe 订阅事件
func (b *eventBus) Subscribe(eventType EventType, handler Handler) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return ""
	}

	b.nextID++
	id := fmt.Sprintf("sub_%d", b.nextID)

	sub := subscription{
		id:      id,
		handler: handler,
	}

	b.subscriptions[eventType] = append(b.subscriptions[eventType], sub)

	return id
}

// Unsubscribe 取消订阅
func (b *eventBus) Unsubscribe(subscriptionID string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("event bus is closed")
	}

	for eventType, subs := range b.subscriptions {
		for i, sub := range subs {
			if sub.id == subscriptionID {
				b.subscriptions[eventType] = append(subs[:i], subs[i+1:]...)
				return nil
			}
		}
	}

	return fmt.Errorf("subscription %s not found", subscriptionID)
}

// Publish 发布事件
func (b *eventBus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	subs := b.subscriptions[event.Type]
	for _, sub := range subs {
		go sub.handler(event)
	}
}

// Close 关闭事件总线
func (b *eventBus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.closed = true
	b.subscriptions = make(map[EventType][]subscription)
}

// 全局事件总线
var defaultBus = NewEventBus()

// Subscribe 使用默认事件总线订阅事件
func Subscribe(eventType EventType, handler Handler) string {
	return defaultBus.Subscribe(eventType, handler)
}

// Unsubscribe 使用默认事件总线取消订阅
func Unsubscribe(subscriptionID string) error {
	return defaultBus.Unsubscribe(subscriptionID)
}

// Publish 使用默认事件总线发布事件
func Publish(event Event) {
	defaultBus.Publish(event)
}

// Close 关闭默认事件总线
func Close() {
	defaultBus.Close()
}
