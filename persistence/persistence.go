// Package persistence 提供了工作流持久化的实现。
// 支持工作流、节点和检查点的存储和加载。
package persistence

import (
	"fmt"
	"sync"
	"time"

	"github.com/ELSTHub/elstworkflow/core"
)

// Store 定义存储接口
type Store interface {
	// SaveWorkflow 保存工作流
	SaveWorkflow(id string, data []byte) error
	// LoadWorkflow 加载工作流
	LoadWorkflow(id string) ([]byte, error)
	// DeleteWorkflow 删除工作流
	DeleteWorkflow(id string) error
	// ListWorkflows 列出所有工作流
	ListWorkflows() ([]string, error)

	// SaveNode 保存节点状态
	SaveNode(workflowID, nodeName string, data []byte) error
	// LoadNode 加载节点状态
	LoadNode(workflowID, nodeName string) ([]byte, error)
	// DeleteNode 删除节点状态
	DeleteNode(workflowID, nodeName string) error
	// ListNodes 列出工作流的所有节点
	ListNodes(workflowID string) ([]string, error)

	// SaveCheckpoint 保存检查点
	SaveCheckpoint(workflowID string, checkpoint *Checkpoint) error
	// LoadCheckpoint 加载检查点
	LoadCheckpoint(workflowID string) (*Checkpoint, error)
	// DeleteCheckpoint 删除检查点
	DeleteCheckpoint(workflowID string) error
}

// Checkpoint 表示工作流检查点
type Checkpoint struct {
	// WorkflowID 工作流ID
	WorkflowID string
	// Status 工作流状态
	Status core.WorkflowStatus
	// CompletedNodes 已完成的节点
	CompletedNodes []string
	// FailedNodes 已失败的节点
	FailedNodes []string
	// NodeResults 节点执行结果
	NodeResults map[string]*core.NodeResult
	// ContextData 上下文数据
	ContextData map[string]interface{}
	// CreatedAt 创建时间
	CreatedAt time.Time
	// UpdatedAt 更新时间
	UpdatedAt time.Time
}

// MemoryStore 内存存储实现
type MemoryStore struct {
	mu          sync.RWMutex
	workflows   map[string][]byte
	nodes       map[string]map[string][]byte
	checkpoints map[string]*Checkpoint
}

// NewMemoryStore 创建新的内存存储
func NewMemoryStore() Store {
	return &MemoryStore{
		workflows:   make(map[string][]byte),
		nodes:       make(map[string]map[string][]byte),
		checkpoints: make(map[string]*Checkpoint),
	}
}

// SaveWorkflow 保存工作流
func (s *MemoryStore) SaveWorkflow(id string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.workflows[id] = data
	return nil
}

// LoadWorkflow 加载工作流
func (s *MemoryStore) LoadWorkflow(id string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.workflows[id]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", id)
	}

	return data, nil
}

// DeleteWorkflow 删除工作流
func (s *MemoryStore) DeleteWorkflow(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.workflows, id)
	delete(s.nodes, id)
	delete(s.checkpoints, id)
	return nil
}

// ListWorkflows 列出所有工作流
func (s *MemoryStore) ListWorkflows() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.workflows))
	for id := range s.workflows {
		ids = append(ids, id)
	}
	return ids, nil
}

// SaveNode 保存节点状态
func (s *MemoryStore) SaveNode(workflowID, nodeName string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.nodes[workflowID] == nil {
		s.nodes[workflowID] = make(map[string][]byte)
	}

	s.nodes[workflowID][nodeName] = data
	return nil
}

// LoadNode 加载节点状态
func (s *MemoryStore) LoadNode(workflowID, nodeName string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes, ok := s.nodes[workflowID]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	data, ok := nodes[nodeName]
	if !ok {
		return nil, fmt.Errorf("node %s not found in workflow %s", nodeName, workflowID)
	}

	return data, nil
}

// DeleteNode 删除节点状态
func (s *MemoryStore) DeleteNode(workflowID, nodeName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	nodes, ok := s.nodes[workflowID]
	if !ok {
		return nil
	}

	delete(nodes, nodeName)
	return nil
}

// ListNodes 列出工作流的所有节点
func (s *MemoryStore) ListNodes(workflowID string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes, ok := s.nodes[workflowID]
	if !ok {
		return nil, fmt.Errorf("workflow %s not found", workflowID)
	}

	names := make([]string, 0, len(nodes))
	for name := range nodes {
		names = append(names, name)
	}
	return names, nil
}

// SaveCheckpoint 保存检查点
func (s *MemoryStore) SaveCheckpoint(workflowID string, checkpoint *Checkpoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	checkpoint.UpdatedAt = time.Now()
	if checkpoint.CreatedAt.IsZero() {
		checkpoint.CreatedAt = checkpoint.UpdatedAt
	}

	s.checkpoints[workflowID] = checkpoint
	return nil
}

// LoadCheckpoint 加载检查点
func (s *MemoryStore) LoadCheckpoint(workflowID string) (*Checkpoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	checkpoint, ok := s.checkpoints[workflowID]
	if !ok {
		return nil, fmt.Errorf("checkpoint for workflow %s not found", workflowID)
	}

	return checkpoint, nil
}

// DeleteCheckpoint 删除检查点
func (s *MemoryStore) DeleteCheckpoint(workflowID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.checkpoints, workflowID)
	return nil
}

// FileStore 文件存储实现（预留）
type FileStore struct {
	basePath string
}

// NewFileStore 创建新的文件存储
func NewFileStore(basePath string) Store {
	return &FileStore{
		basePath: basePath,
	}
}

// SaveWorkflow 保存工作流
func (s *FileStore) SaveWorkflow(id string, data []byte) error {
	// 预留实现
	return fmt.Errorf("not implemented")
}

// LoadWorkflow 加载工作流
func (s *FileStore) LoadWorkflow(id string) ([]byte, error) {
	// 预留实现
	return nil, fmt.Errorf("not implemented")
}

// DeleteWorkflow 删除工作流
func (s *FileStore) DeleteWorkflow(id string) error {
	// 预留实现
	return fmt.Errorf("not implemented")
}

// ListWorkflows 列出所有工作流
func (s *FileStore) ListWorkflows() ([]string, error) {
	// 预留实现
	return nil, fmt.Errorf("not implemented")
}

// SaveNode 保存节点状态
func (s *FileStore) SaveNode(workflowID, nodeName string, data []byte) error {
	// 预留实现
	return fmt.Errorf("not implemented")
}

// LoadNode 加载节点状态
func (s *FileStore) LoadNode(workflowID, nodeName string) ([]byte, error) {
	// 预留实现
	return nil, fmt.Errorf("not implemented")
}

// DeleteNode 删除节点状态
func (s *FileStore) DeleteNode(workflowID, nodeName string) error {
	// 预留实现
	return fmt.Errorf("not implemented")
}

// ListNodes 列出工作流的所有节点
func (s *FileStore) ListNodes(workflowID string) ([]string, error) {
	// 预留实现
	return nil, fmt.Errorf("not implemented")
}

// SaveCheckpoint 保存检查点
func (s *FileStore) SaveCheckpoint(workflowID string, checkpoint *Checkpoint) error {
	// 预留实现
	return fmt.Errorf("not implemented")
}

// LoadCheckpoint 加载检查点
func (s *FileStore) LoadCheckpoint(workflowID string) (*Checkpoint, error) {
	// 预留实现
	return nil, fmt.Errorf("not implemented")
}

// DeleteCheckpoint 删除检查点
func (s *FileStore) DeleteCheckpoint(workflowID string) error {
	// 预留实现
	return fmt.Errorf("not implemented")
}
