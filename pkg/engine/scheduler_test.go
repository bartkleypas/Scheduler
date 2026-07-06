package engine

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"scheduler/pkg/parser"
)

// mockGraphServiceSpy is a spy to check what calls were made to the GraphService interface.
type mockGraphServiceSpy struct {
	mu           sync.Mutex
	saveCalls    []*parser.Graph
	deleteCalls  []string
	closeCalled  bool
	getGraphFunc func() (*parser.Graph, error)
}

func (m *mockGraphServiceSpy) SaveGraph(ctx context.Context, g *parser.Graph) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveCalls = append(m.saveCalls, g)
	return nil
}

func (m *mockGraphServiceSpy) GetGraph(ctx context.Context) (*parser.Graph, error) {
	if m.getGraphFunc != nil {
		return m.getGraphFunc()
	}
	return nil, nil
}

func (m *mockGraphServiceSpy) DeleteNode(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalls = append(m.deleteCalls, id)
	return nil
}

func (m *mockGraphServiceSpy) DeleteFileGraph(ctx context.Context, relPath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deleteCalls = append(m.deleteCalls, relPath)
	return nil
}

func (m *mockGraphServiceSpy) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalled = true
	return nil
}

func TestNewSchedulerDefaults(t *testing.T) {
	cfg := Config{}
	s := NewScheduler(cfg, &mockGraphServiceSpy{})
	if s.config.WatchDebounce != 500*time.Millisecond {
		t.Errorf("expected default WatchDebounce to be 500ms, got %s", s.config.WatchDebounce)
	}
}

func TestScheduler_WallClockTicker(t *testing.T) {
	cfg := Config{
		ScanInterval: 10 * time.Millisecond,
	}
	s := NewScheduler(cfg, &mockGraphServiceSpy{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.runWallClock(ctx)

	// We expect multiple ticks to be sent to s.triggers
	select {
	case trigger := <-s.triggers:
		if trigger.Type != TriggerPeriodicScan {
			t.Errorf("expected trigger type %s, got %s", TriggerPeriodicScan, trigger.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timed out waiting for periodic scan trigger")
	}
}

func TestScheduler_IngestUnsupportedFile(t *testing.T) {
	tempWorkspace := t.TempDir()
	cfg := Config{
		WorkspacePath: tempWorkspace,
	}

	spy := &mockGraphServiceSpy{}
	s := NewScheduler(cfg, spy)

	// Trigger ingest for a txt file (unsupported extension)
	s.ingest(IngestTrigger{
		Type: TriggerFileChanged,
		Path: filepath.Join(tempWorkspace, "readme.txt"),
	})

	// SaveGraph should not have been called because the extension is unsupported
	spy.mu.Lock()
	defer spy.mu.Unlock()
	if len(spy.saveCalls) != 0 {
		t.Errorf("expected SaveGraph not to be called, but got %d calls", len(spy.saveCalls))
	}
}

func TestScheduler_IngestIncremental(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Write mock Go file
	goFile := filepath.Join(tempWorkspace, "main.go")
	goSource := `package main
	func main() {}`
	if err := os.WriteFile(goFile, []byte(goSource), 0644); err != nil {
		t.Fatalf("failed to write mock Go file: %v", err)
	}

	cfg := Config{
		WorkspacePath: tempWorkspace,
	}

	spy := &mockGraphServiceSpy{}
	s := NewScheduler(cfg, spy)

	// Trigger ingest for the Go file
	s.ingest(IngestTrigger{
		Type: TriggerFileChanged,
		Path: goFile,
	})

	spy.mu.Lock()
	defer spy.mu.Unlock()
	if len(spy.saveCalls) != 1 {
		t.Fatalf("expected SaveGraph to be called once, got %d calls", len(spy.saveCalls))
	}

	g := spy.saveCalls[0]
	// Verify that the graph contains elements from the Go file
	if _, ok := g.Nodes["main.main"]; !ok {
		t.Error("expected Go symbol node main.main not found in incremental ingest graph")
	}
}

func TestScheduler_IngestFullScan(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Write mock Go file
	goFile := filepath.Join(tempWorkspace, "main.go")
	goSource := `package main
	func main() {}`
	if err := os.WriteFile(goFile, []byte(goSource), 0644); err != nil {
		t.Fatalf("failed to write mock Go file: %v", err)
	}

	cfg := Config{
		WorkspacePath: tempWorkspace,
	}

	spy := &mockGraphServiceSpy{}
	s := NewScheduler(cfg, spy)

	// Trigger full scan
	s.ingest(IngestTrigger{
		Type: TriggerPeriodicScan,
	})

	spy.mu.Lock()
	defer spy.mu.Unlock()
	if len(spy.saveCalls) != 1 {
		t.Fatalf("expected SaveGraph to be called once, got %d calls", len(spy.saveCalls))
	}

	g := spy.saveCalls[0]
	if _, ok := g.Nodes["main.main"]; !ok {
		t.Error("expected Go symbol node main.main in full scan graph")
	}
}
