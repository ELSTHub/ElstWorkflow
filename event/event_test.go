package event

import (
	"sync"
	"testing"
	"time"
)

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus()
	if bus == nil {
		t.Fatal("NewEventBus() returned nil")
	}
}

func TestSubscribe(t *testing.T) {
	bus := NewEventBus()

	handler := func(event Event) {}
	id := bus.Subscribe(WorkflowStarted, handler)

	if id == "" {
		t.Error("expected non-empty subscription ID")
	}
}

func TestPublish(t *testing.T) {
	bus := NewEventBus()

	var received Event
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event Event) {
		received = event
		wg.Done()
	}

	bus.Subscribe(WorkflowStarted, handler)

	event := Event{
		Type:         WorkflowStarted,
		WorkflowName: "test-workflow",
		Timestamp:    time.Now(),
	}

	bus.Publish(event)
	wg.Wait()

	if received.Type != WorkflowStarted {
		t.Errorf("expected type WorkflowStarted, got %v", received.Type)
	}
	if received.WorkflowName != "test-workflow" {
		t.Errorf("expected workflow name 'test-workflow', got '%s'", received.WorkflowName)
	}
}

func TestPublishMultipleHandlers(t *testing.T) {
	bus := NewEventBus()

	var count int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		bus.Subscribe(NodeStarted, func(event Event) {
			mu.Lock()
			count++
			mu.Unlock()
			wg.Done()
		})
	}

	bus.Publish(Event{Type: NodeStarted})
	wg.Wait()

	if count != 3 {
		t.Errorf("expected 3 handlers called, got %d", count)
	}
}

func TestUnsubscribe(t *testing.T) {
	bus := NewEventBus()

	called := false
	handler := func(event Event) {
		called = true
	}

	id := bus.Subscribe(WorkflowStarted, handler)

	if err := bus.Unsubscribe(id); err != nil {
		t.Errorf("Unsubscribe failed: %v", err)
	}

	bus.Publish(Event{Type: WorkflowStarted})
	time.Sleep(10 * time.Millisecond)

	if called {
		t.Error("handler should not have been called")
	}
}

func TestUnsubscribeNotFound(t *testing.T) {
	bus := NewEventBus()

	if err := bus.Unsubscribe("nonexistent"); err == nil {
		t.Error("expected error for nonexistent subscription")
	}
}

func TestClose(t *testing.T) {
	bus := NewEventBus()

	called := false
	handler := func(event Event) {
		called = true
	}

	bus.Subscribe(WorkflowStarted, handler)
	bus.Close()

	bus.Publish(Event{Type: WorkflowStarted})
	time.Sleep(10 * time.Millisecond)

	if called {
		t.Error("handler should not have been called after close")
	}
}

func TestPublishAfterClose(t *testing.T) {
	bus := NewEventBus()
	bus.Close()

	// 应该不会panic
	bus.Publish(Event{Type: WorkflowStarted})
}

func TestSubscribeAfterClose(t *testing.T) {
	bus := NewEventBus()
	bus.Close()

	id := bus.Subscribe(WorkflowStarted, func(event Event) {})
	if id != "" {
		t.Error("expected empty subscription ID after close")
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{WorkflowStarted, "WorkflowStarted"},
		{WorkflowFinished, "WorkflowFinished"},
		{WorkflowFailed, "WorkflowFailed"},
		{WorkflowCancelled, "WorkflowCancelled"},
		{NodeStarted, "NodeStarted"},
		{NodeFinished, "NodeFinished"},
		{NodeFailed, "NodeFailed"},
		{NodeRetry, "NodeRetry"},
		{NodeRollback, "NodeRollback"},
	}

	for _, tt := range tests {
		if tt.eventType.String() != tt.expected {
			t.Errorf("expected %s, got %s", tt.expected, tt.eventType.String())
		}
	}
}

func TestEventBusInterface(t *testing.T) {
	// 确保 eventBus 实现了 EventBus 接口
	var _ EventBus = NewEventBus()
}

func TestGlobalEventBus(t *testing.T) {
	// 测试全局事件总线
	var received bool
	var wg sync.WaitGroup
	wg.Add(1)

	handler := func(event Event) {
		received = true
		wg.Done()
	}

	id := Subscribe(WorkflowStarted, handler)
	defer Unsubscribe(id)

	Publish(Event{Type: WorkflowStarted})
	wg.Wait()

	if !received {
		t.Error("expected event to be received")
	}
}

// 基准测试
func BenchmarkPublish(b *testing.B) {
	bus := NewEventBus()
	bus.Subscribe(WorkflowStarted, func(event Event) {})
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.Publish(Event{Type: WorkflowStarted})
	}
}

func BenchmarkPublishMultipleHandlers(b *testing.B) {
	bus := NewEventBus()
	for i := 0; i < 10; i++ {
		bus.Subscribe(WorkflowStarted, func(event Event) {})
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bus.Publish(Event{Type: WorkflowStarted})
	}
}
