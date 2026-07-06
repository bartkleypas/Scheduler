package testutil

import (
	"context"
	"scheduler/pkg/parser"
	"sync"
)

// MockGraphService is a thread-safe unified spy/stub for GraphService testing.
type MockGraphService struct {
	mu              sync.Mutex
	Graph           *parser.Graph
	DeleteCalls     []string
	DeleteFileCalls []string
	SaveCalls       []*parser.Graph
	CloseCalled     bool
	FileHashes      map[string]string
	GetGraphFunc    func() (*parser.Graph, error)
}

// NewMockGraphService creates a new mock graph service instance.
func NewMockGraphService() *MockGraphService {
	return &MockGraphService{
		FileHashes: make(map[string]string),
	}
}

func (m *MockGraphService) SaveGraph(ctx context.Context, g *parser.Graph) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SaveCalls = append(m.SaveCalls, g)
	m.Graph = g
	return nil
}

func (m *MockGraphService) GetGraph(ctx context.Context) (*parser.Graph, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.GetGraphFunc != nil {
		return m.GetGraphFunc()
	}
	return m.Graph, nil
}

func (m *MockGraphService) DeleteNode(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteCalls = append(m.DeleteCalls, id)
	return nil
}

func (m *MockGraphService) DeleteFileGraph(ctx context.Context, relPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DeleteFileCalls = append(m.DeleteFileCalls, relPath)
	return nil
}

func (m *MockGraphService) GetFileHash(ctx context.Context, relPath string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.FileHashes[relPath], nil
}

func (m *MockGraphService) SetFileHash(ctx context.Context, relPath, hash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.FileHashes == nil {
		m.FileHashes = make(map[string]string)
	}
	m.FileHashes[relPath] = hash
	return nil
}

func (m *MockGraphService) DeleteFileHash(ctx context.Context, relPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.FileHashes, relPath)
	return nil
}

func (m *MockGraphService) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.CloseCalled = true
	return nil
}
