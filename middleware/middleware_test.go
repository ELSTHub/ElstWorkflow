package middleware

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ELSTHub/elstworkflow/context"
	"github.com/ELSTHub/elstworkflow/core"
	"github.com/ELSTHub/elstworkflow/node"
)

func TestNewChain(t *testing.T) {
	chain := NewChain()
	if chain == nil {
		t.Fatal("NewChain() returned nil")
	}
}

func TestChainAdd(t *testing.T) {
	chain := NewChain()
	chain.Add(DefaultLoggingMiddleware())

	if len(chain.middlewares) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(chain.middlewares))
	}
}

func TestChainExecute(t *testing.T) {
	beforeCalled := false
	afterCalled := false

	m := &testMiddleware{
		before: func(ctx *Context, nodeCtx core.Context, n node.Node) error {
			beforeCalled = true
			return nil
		},
		after: func(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error) {
			afterCalled = true
		},
	}

	chain := NewChain(m)
	nodeCtx := context.New()
	n := node.NewFuncNode("test", func(ctx core.Context) (interface{}, error) {
		return "result", nil
	})

	result, err := chain.Execute("workflow", "test", nodeCtx, n, func() (interface{}, error) {
		return "result", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "result" {
		t.Errorf("expected 'result', got %v", result)
	}
	if !beforeCalled {
		t.Error("Before not called")
	}
	if !afterCalled {
		t.Error("After not called")
	}
}

func TestChainExecuteError(t *testing.T) {
	onErrorCalled := false

	m := &testMiddleware{
		onError: func(ctx *Context, nodeCtx core.Context, n node.Node, err error) {
			onErrorCalled = true
		},
	}

	chain := NewChain(m)
	nodeCtx := context.New()
	n := node.NewFuncNode("test", func(ctx core.Context) (interface{}, error) {
		return nil, errors.New("test error")
	})

	_, err := chain.Execute("workflow", "test", nodeCtx, n, func() (interface{}, error) {
		return nil, errors.New("test error")
	})

	if err == nil {
		t.Error("expected error")
	}
	if !onErrorCalled {
		t.Error("OnError not called")
	}
}

func TestLoggingMiddleware(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	m := NewLoggingMiddleware(logger)

	ctx := &Context{
		WorkflowName: "test",
		NodeName:     "node1",
		StartTime:    time.Now(),
	}
	nodeCtx := context.New()
	n := node.NewFuncNode("node1", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	if err := m.Before(ctx, nodeCtx, n); err != nil {
		t.Errorf("Before failed: %v", err)
	}

	m.After(ctx, nodeCtx, n, nil, nil)
	m.OnError(ctx, nodeCtx, n, errors.New("test"))
}

func TestMetricsMiddleware(t *testing.T) {
	m := NewMetricsMiddleware()

	ctx := &Context{
		StartTime: time.Now(),
	}
	nodeCtx := context.New()
	n := node.NewFuncNode("node1", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	m.Before(ctx, nodeCtx, n)
	m.After(ctx, nodeCtx, n, nil, nil)

	if m.TotalRequests != 1 {
		t.Errorf("expected 1 request, got %d", m.TotalRequests)
	}
	if m.SuccessCount != 1 {
		t.Errorf("expected 1 success, got %d", m.SuccessCount)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	recovered := false
	m := NewRecoveryMiddleware(func(ctx *Context, nodeCtx core.Context, n node.Node, r interface{}) {
		recovered = true
	})

	ctx := &Context{}
	nodeCtx := context.New()
	n := node.NewFuncNode("node1", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	m.Before(ctx, nodeCtx, n)
	m.After(ctx, nodeCtx, n, nil, nil)
	m.OnError(ctx, nodeCtx, n, errors.New("test"))

	if !recovered {
		t.Error("recovery handler not called")
	}
}

func TestAuditMiddleware(t *testing.T) {
	auditCalled := false
	m := NewAuditMiddleware(func(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error, duration time.Duration) {
		auditCalled = true
	})

	ctx := &Context{StartTime: time.Now()}
	nodeCtx := context.New()
	n := node.NewFuncNode("node1", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	m.Before(ctx, nodeCtx, n)
	m.After(ctx, nodeCtx, n, nil, nil)

	if !auditCalled {
		t.Error("audit handler not called")
	}
}

func TestTracingMiddleware(t *testing.T) {
	traces := make([]string, 0)
	m := NewTracingMiddleware(func(ctx *Context, nodeCtx core.Context, n node.Node, phase string, duration time.Duration) {
		traces = append(traces, phase)
	})

	ctx := &Context{StartTime: time.Now()}
	nodeCtx := context.New()
	n := node.NewFuncNode("node1", func(ctx core.Context) (interface{}, error) {
		return nil, nil
	})

	m.Before(ctx, nodeCtx, n)
	m.After(ctx, nodeCtx, n, nil, nil)

	if len(traces) != 2 {
		t.Errorf("expected 2 traces, got %d", len(traces))
	}
}

func TestMiddlewareInterface(t *testing.T) {
	// 确保所有中间件类型都实现了Middleware接口
	var _ Middleware = NewLoggingMiddleware(nil)
	var _ Middleware = NewMetricsMiddleware()
	var _ Middleware = NewRecoveryMiddleware(nil)
	var _ Middleware = NewAuditMiddleware(nil)
	var _ Middleware = NewTracingMiddleware(nil)
}

// 测试中间件
type testMiddleware struct {
	before  func(ctx *Context, nodeCtx core.Context, n node.Node) error
	after   func(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error)
	onError func(ctx *Context, nodeCtx core.Context, n node.Node, err error)
}

func (m *testMiddleware) Before(ctx *Context, nodeCtx core.Context, n node.Node) error {
	if m.before != nil {
		return m.before(ctx, nodeCtx, n)
	}
	return nil
}

func (m *testMiddleware) After(ctx *Context, nodeCtx core.Context, n node.Node, result interface{}, err error) {
	if m.after != nil {
		m.after(ctx, nodeCtx, n, result, err)
	}
}

func (m *testMiddleware) OnError(ctx *Context, nodeCtx core.Context, n node.Node, err error) {
	if m.onError != nil {
		m.onError(ctx, nodeCtx, n, err)
	}
}
