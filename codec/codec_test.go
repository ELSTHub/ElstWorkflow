package codec

import (
	"testing"
)

func TestNewJSONCodec(t *testing.T) {
	codec := NewJSONCodec("  ")
	if codec == nil {
		t.Fatal("NewJSONCodec() returned nil")
	}
	if codec.ContentType() != "application/json" {
		t.Errorf("expected 'application/json', got '%s'", codec.ContentType())
	}
}

func TestJSONCodecMarshal(t *testing.T) {
	codec := NewJSONCodec("")

	data := map[string]string{
		"key": "value",
	}

	bytes, err := codec.Marshal(data)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}
	if len(bytes) == 0 {
		t.Error("expected non-empty bytes")
	}
}

func TestJSONCodecMarshalIndent(t *testing.T) {
	codec := NewJSONCodec("  ")

	data := map[string]string{
		"key": "value",
	}

	bytes, err := codec.Marshal(data)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}
	if len(bytes) == 0 {
		t.Error("expected non-empty bytes")
	}
}

func TestJSONCodecUnmarshal(t *testing.T) {
	codec := NewJSONCodec("")

	data := []byte(`{"key":"value"}`)
	var result map[string]string

	if err := codec.Unmarshal(data, &result); err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected 'value', got '%s'", result["key"])
	}
}

func TestNewYAMLCodec(t *testing.T) {
	codec := NewYAMLCodec()
	if codec == nil {
		t.Fatal("NewYAMLCodec() returned nil")
	}
	if codec.ContentType() != "application/yaml" {
		t.Errorf("expected 'application/yaml', got '%s'", codec.ContentType())
	}
}

func TestYAMLCodecMarshal(t *testing.T) {
	codec := NewYAMLCodec()

	data := map[string]string{
		"key": "value",
	}

	bytes, err := codec.Marshal(data)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}
	if len(bytes) == 0 {
		t.Error("expected non-empty bytes")
	}
}

func TestYAMLCodecUnmarshal(t *testing.T) {
	codec := NewYAMLCodec()

	data := []byte(`{"key":"value"}`)
	var result map[string]string

	if err := codec.Unmarshal(data, &result); err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected 'value', got '%s'", result["key"])
	}
}

func TestMarshalJSON(t *testing.T) {
	data := map[string]string{
		"key": "value",
	}

	bytes, err := MarshalJSON(data, "  ")
	if err != nil {
		t.Errorf("MarshalJSON failed: %v", err)
	}
	if len(bytes) == 0 {
		t.Error("expected non-empty bytes")
	}
}

func TestUnmarshalJSON(t *testing.T) {
	data := []byte(`{"key":"value"}`)
	var result map[string]string

	if err := UnmarshalJSON(data, &result); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected 'value', got '%s'", result["key"])
	}
}

func TestMarshalYAML(t *testing.T) {
	data := map[string]string{
		"key": "value",
	}

	bytes, err := MarshalYAML(data)
	if err != nil {
		t.Errorf("MarshalYAML failed: %v", err)
	}
	if len(bytes) == 0 {
		t.Error("expected non-empty bytes")
	}
}

func TestUnmarshalYAML(t *testing.T) {
	data := []byte(`{"key":"value"}`)
	var result map[string]string

	if err := UnmarshalYAML(data, &result); err != nil {
		t.Errorf("UnmarshalYAML failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("expected 'value', got '%s'", result["key"])
	}
}

func TestWorkflowData(t *testing.T) {
	data := &WorkflowData{
		Name:        "test-workflow",
		Version:     "1.0.0",
		Description: "Test workflow",
		Nodes: []NodeData{
			{Name: "node1", Type: "func"},
			{Name: "node2", Type: "func"},
		},
		Edges: []EdgeData{
			{From: "node1", To: "node2"},
		},
		Metadata: map[string]string{
			"env": "test",
		},
	}

	bytes, err := MarshalJSON(data, "  ")
	if err != nil {
		t.Errorf("MarshalJSON failed: %v", err)
	}

	var decoded WorkflowData
	if err := UnmarshalJSON(bytes, &decoded); err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if decoded.Name != "test-workflow" {
		t.Errorf("expected 'test-workflow', got '%s'", decoded.Name)
	}
	if len(decoded.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(decoded.Nodes))
	}
	if len(decoded.Edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(decoded.Edges))
	}
}

func TestValidate(t *testing.T) {
	// 有效数据
	validData := &WorkflowData{
		Name: "test",
		Nodes: []NodeData{
			{Name: "node1"},
			{Name: "node2"},
		},
		Edges: []EdgeData{
			{From: "node1", To: "node2"},
		},
	}

	if err := Validate(validData); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// 无效数据：空名称
	invalidData1 := &WorkflowData{
		Nodes: []NodeData{{Name: "node1"}},
	}
	if err := Validate(invalidData1); err == nil {
		t.Error("expected error for empty name")
	}

	// 无效数据：无节点
	invalidData2 := &WorkflowData{
		Name: "test",
	}
	if err := Validate(invalidData2); err == nil {
		t.Error("expected error for no nodes")
	}

	// 无效数据：重复节点名
	invalidData3 := &WorkflowData{
		Name: "test",
		Nodes: []NodeData{
			{Name: "node1"},
			{Name: "node1"},
		},
	}
	if err := Validate(invalidData3); err == nil {
		t.Error("expected error for duplicate node names")
	}

	// 无效数据：边引用不存在的节点
	invalidData4 := &WorkflowData{
		Name: "test",
		Nodes: []NodeData{
			{Name: "node1"},
		},
		Edges: []EdgeData{
			{From: "node1", To: "nonexistent"},
		},
	}
	if err := Validate(invalidData4); err == nil {
		t.Error("expected error for nonexistent edge target")
	}
}

func TestCodecInterface(t *testing.T) {
	// 确保所有编解码器类型都实现了Codec接口
	var _ Codec = NewJSONCodec("")
	var _ Codec = NewYAMLCodec()
}

// 基准测试
func BenchmarkJSONMarshal(b *testing.B) {
	codec := NewJSONCodec("")
	data := map[string]string{"key": "value"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		codec.Marshal(data)
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	codec := NewJSONCodec("")
	data := []byte(`{"key":"value"}`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result map[string]string
		codec.Unmarshal(data, &result)
	}
}
