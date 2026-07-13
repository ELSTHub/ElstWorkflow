// Package codec 提供了工作流编解码器的实现。
// 支持 JSON 和 YAML 格式的工作流序列化和反序列化。
package codec

import (
	"encoding/json"
	"fmt"
)

// Codec 定义编解码器接口
type Codec interface {
	// Marshal 序列化对象为字节
	Marshal(v interface{}) ([]byte, error)
	// Unmarshal 反序列化字节为对象
	Unmarshal(data []byte, v interface{}) error
	// ContentType 返回内容类型
	ContentType() string
}

// JSONCodec JSON编解码器
type JSONCodec struct {
	indent string
}

// NewJSONCodec 创建JSON编解码器
func NewJSONCodec(indent string) *JSONCodec {
	return &JSONCodec{
		indent: indent,
	}
}

// Marshal 序列化对象为JSON字节
func (c *JSONCodec) Marshal(v interface{}) ([]byte, error) {
	if c.indent != "" {
		return json.MarshalIndent(v, "", c.indent)
	}
	return json.Marshal(v)
}

// Unmarshal 反序列化JSON字节为对象
func (c *JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// ContentType 返回内容类型
func (c *JSONCodec) ContentType() string {
	return "application/json"
}

// YAMLCodec YAML编解码器（预留实现）
type YAMLCodec struct{}

// NewYAMLCodec 创建YAML编解码器
func NewYAMLCodec() *YAMLCodec {
	return &YAMLCodec{}
}

// Marshal 序列化对象为YAML字节
func (c *YAMLCodec) Marshal(v interface{}) ([]byte, error) {
	// 先转换为JSON，然后转换为YAML格式
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// 简单的JSON到YAML转换（生产环境应使用专用YAML库）
	var m map[string]interface{}
	if err := json.Unmarshal(jsonData, &m); err != nil {
		return nil, err
	}

	return mapToYAML(m, 0), nil
}

// Unmarshal 反序列化YAML字节为对象
func (c *YAMLCodec) Unmarshal(data []byte, v interface{}) error {
	// 简单的YAML到JSON转换（生产环境应使用专用YAML库）
	jsonData, err := yamlToJSON(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, v)
}

// ContentType 返回内容类型
func (c *YAMLCodec) ContentType() string {
	return "application/yaml"
}

// mapToYAML 将map转换为YAML格式
func mapToYAML(m map[string]interface{}, indent int) []byte {
	result := make([]byte, 0)
	indentStr := ""
	for i := 0; i < indent; i++ {
		indentStr += "  "
	}

	for key, value := range m {
		result = append(result, []byte(fmt.Sprintf("%s%s:", indentStr, key))...)
		result = append(result, '\n')

		switch v := value.(type) {
		case map[string]interface{}:
			result = append(result, mapToYAML(v, indent+1)...)
		case string:
			result = append(result, []byte(fmt.Sprintf("%s  %s\n", indentStr, v))...)
		default:
			result = append(result, []byte(fmt.Sprintf("%s  %v\n", indentStr, v))...)
		}
	}

	return result
}

// yamlToJSON 将YAML转换为JSON
func yamlToJSON(data []byte) ([]byte, error) {
	// 简单实现：假设输入已经是JSON格式
	// 生产环境应使用专用YAML库
	return data, nil
}

// WorkflowData 表示工作流的可序列化数据
type WorkflowData struct {
	// Name 工作流名称
	Name string `json:"name" yaml:"name"`
	// Version 工作流版本
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// Description 工作流描述
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Nodes 节点列表
	Nodes []NodeData `json:"nodes" yaml:"nodes"`
	// Edges 边列表
	Edges []EdgeData `json:"edges,omitempty" yaml:"edges,omitempty"`
	// Metadata 元数据
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// NodeData 表示节点的可序列化数据
type NodeData struct {
	// Name 节点名称
	Name string `json:"name" yaml:"name"`
	// Type 节点类型
	Type string `json:"type" yaml:"type"`
	// Description 节点描述
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Metadata 元数据
	Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// EdgeData 表示边的可序列化数据
type EdgeData struct {
	// From 源节点名称
	From string `json:"from" yaml:"from"`
	// To 目标节点名称
	To string `json:"to" yaml:"to"`
}

// MarshalJSON 序列化工作流为JSON
func MarshalJSON(v interface{}, indent string) ([]byte, error) {
	codec := NewJSONCodec(indent)
	return codec.Marshal(v)
}

// UnmarshalJSON 反序列化JSON为对象
func UnmarshalJSON(data []byte, v interface{}) error {
	codec := NewJSONCodec("")
	return codec.Unmarshal(data, v)
}

// MarshalYAML 序列化工作流为YAML
func MarshalYAML(v interface{}) ([]byte, error) {
	codec := NewYAMLCodec()
	return codec.Marshal(v)
}

// UnmarshalYAML 反序列化YAML为对象
func UnmarshalYAML(data []byte, v interface{}) error {
	codec := NewYAMLCodec()
	return codec.Unmarshal(data, v)
}

// Validate 验证工作流数据
func Validate(data *WorkflowData) error {
	if data.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(data.Nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	// 检查节点名称是否唯一
	nodeNames := make(map[string]bool)
	for _, n := range data.Nodes {
		if n.Name == "" {
			return fmt.Errorf("node name is required")
		}
		if nodeNames[n.Name] {
			return fmt.Errorf("duplicate node name: %s", n.Name)
		}
		nodeNames[n.Name] = true
	}

	// 检查边的节点是否存在
	for _, e := range data.Edges {
		if !nodeNames[e.From] {
			return fmt.Errorf("edge source node %s not found", e.From)
		}
		if !nodeNames[e.To] {
			return fmt.Errorf("edge target node %s not found", e.To)
		}
	}

	return nil
}
