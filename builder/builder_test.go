package builder

import (
	"testing"

	"github.com/ELSTHub/elstworkflow/core"
)

func TestNewBuilder(t *testing.T) {
	b := New("test-workflow")
	if b == nil {
		t.Fatal("New() returned nil")
	}
}

func TestBuildSimpleWorkflow(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return "result1", nil
		}).
		Node("node2", func(ctx core.Context) (interface{}, error) {
			return "result2", nil
		}).
		DependsOn("node2", "node1").
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithRollback(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		NodeWithRollback("node1",
			func(ctx core.Context) (interface{}, error) {
				return "result1", nil
			},
			func(ctx core.Context) error {
				return nil
			},
		).
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithStaticNode(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		StaticNode("config", map[string]string{
			"key": "value",
		}).
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithDependencies(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Node("node2", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Node("node3", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		DependsOn("node2", "node1").
		DependsOn("node3", "node2").
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithParallel(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Node("node2", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Parallel("node1").
		Parallel("node2").
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithCondition(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Condition("node1", func(ctx core.Context) (bool, error) {
			return true, nil
		}).
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithRetryPolicy(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		WithRetryPolicy("node1", &core.RetryPolicy{
			Strategy:   core.RetryExponential,
			MaxRetries: 3,
		}).
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithTimeoutPolicy(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		WithTimeoutPolicy("node1", &core.TimeoutPolicy{
			Timeout: 30,
		}).
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithMetadata(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		WithMetadata("node1", core.Metadata{
			"key1": "value1",
			"key2": "value2",
		}).
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWithVersion(t *testing.T) {
	b := New("test-workflow")

	wf, err := b.
		WithVersion("1.0.0").
		WithDescription("Test workflow").
		WithWorkflowMetadata(core.Metadata{
			"env": "test",
		}).
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	if wf.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", wf.Version)
	}

	if wf.Description != "Test workflow" {
		t.Errorf("expected description 'Test workflow', got '%s'", wf.Description)
	}

	if wf.Metadata["env"] != "test" {
		t.Errorf("expected metadata env='test', got '%s'", wf.Metadata["env"])
	}
}

func TestBuildDuplicateNode(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Build()

	if err == nil {
		t.Error("expected error for duplicate node")
	}
}

func TestBuildNonexistentDependency(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		DependsOn("node1", "nonexistent").
		Build()

	if err == nil {
		t.Error("expected error for nonexistent dependency")
	}
}

func TestBuildNonexistentNodeForOption(t *testing.T) {
	b := New("test-workflow")

	_, err := b.
		Parallel("nonexistent").
		Build()

	if err == nil {
		t.Error("expected error for nonexistent node")
	}
}

func TestBuildWithChain(t *testing.T) {
	b := New("test-workflow")

	deps := Chain("node1", "node2", "node3")

	_, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Node("node2", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		Node("node3", func(ctx core.Context) (interface{}, error) {
			return nil, nil
		}).
		WithDependencies(deps).
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}
}

func TestBuildWorkflowStructure(t *testing.T) {
	b := New("test-workflow")

	wf, err := b.
		Node("node1", func(ctx core.Context) (interface{}, error) {
			return "result1", nil
		}).
		Node("node2", func(ctx core.Context) (interface{}, error) {
			return "result2", nil
		}).
		DependsOn("node2", "node1").
		Build()

	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	if wf.Name != "test-workflow" {
		t.Errorf("expected workflow name 'test-workflow', got '%s'", wf.Name)
	}

	if wf.Graph.NodeCount() != 2 {
		t.Errorf("expected 2 nodes, got %d", wf.Graph.NodeCount())
	}

	if wf.Graph.EdgeCount() != 1 {
		t.Errorf("expected 1 edge, got %d", wf.Graph.EdgeCount())
	}
}

func TestBuilderInterface(t *testing.T) {
	// 确保 workflowBuilder 实现了 Builder 接口
	var _ Builder = New("test")
}

// 基准测试
func BenchmarkBuildSimpleWorkflow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := New("bench-workflow")
		builder.
			Node("node1", func(ctx core.Context) (interface{}, error) {
				return nil, nil
			}).
			Node("node2", func(ctx core.Context) (interface{}, error) {
				return nil, nil
			}).
			DependsOn("node2", "node1").
			Build()
	}
}

func BenchmarkBuildComplexWorkflow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		builder := New("bench-workflow")
		builder.
			Node("node1", func(ctx core.Context) (interface{}, error) {
				return nil, nil
			}).
			Node("node2", func(ctx core.Context) (interface{}, error) {
				return nil, nil
			}).
			Node("node3", func(ctx core.Context) (interface{}, error) {
				return nil, nil
			}).
			Node("node4", func(ctx core.Context) (interface{}, error) {
				return nil, nil
			}).
			DependsOn("node2", "node1").
			DependsOn("node3", "node1").
			DependsOn("node4", "node2", "node3").
			Parallel("node2").
			Parallel("node3").
			Build()
	}
}
