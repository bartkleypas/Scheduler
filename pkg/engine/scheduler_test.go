package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"scheduler/internal/testutil"
)



func TestNewSchedulerDefaults(t *testing.T) {
	cfg := Config{}
	s := NewScheduler(cfg, testutil.NewMockGraphService())
	if s.config.WatchDebounce != 500*time.Millisecond {
		t.Errorf("expected default WatchDebounce to be 500ms, got %s", s.config.WatchDebounce)
	}
}

func TestScheduler_WallClockTicker(t *testing.T) {
	cfg := Config{
		ScanInterval: 10 * time.Millisecond,
	}
	s := NewScheduler(cfg, testutil.NewMockGraphService())

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

	spy := testutil.NewMockGraphService()
	s := NewScheduler(cfg, spy)

	// Trigger ingest for a txt file (unsupported extension)
	s.ingest(IngestTrigger{
		Type: TriggerFileChanged,
		Path: filepath.Join(tempWorkspace, "readme.txt"),
	})

	// SaveGraph should not have been called because the extension is unsupported
	if len(spy.SaveCalls) != 0 {
		t.Errorf("expected SaveGraph not to be called, but got %d calls", len(spy.SaveCalls))
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

	spy := testutil.NewMockGraphService()
	s := NewScheduler(cfg, spy)

	// Trigger ingest for the Go file
	s.ingest(IngestTrigger{
		Type: TriggerFileChanged,
		Path: goFile,
	})

	if len(spy.SaveCalls) != 1 {
		t.Fatalf("expected SaveGraph to be called once, got %d calls", len(spy.SaveCalls))
	}

	g := spy.SaveCalls[0]
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

	spy := testutil.NewMockGraphService()
	s := NewScheduler(cfg, spy)

	// Trigger full scan
	s.ingest(IngestTrigger{
		Type: TriggerPeriodicScan,
	})

	if len(spy.SaveCalls) != 1 {
		t.Fatalf("expected SaveGraph to be called once, got %d calls", len(spy.SaveCalls))
	}

	g := spy.SaveCalls[0]
	if _, ok := g.Nodes["main.main"]; !ok {
		t.Error("expected Go symbol node main.main in full scan graph")
	}
}
