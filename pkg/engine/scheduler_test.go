package engine

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
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

func TestNewScheduler_ResolvesWorkspacePath(t *testing.T) {
	cfg := Config{
		WorkspacePath: ".",
	}
	s := NewScheduler(cfg, testutil.NewMockGraphService())
	if s.config.WorkspacePath == "." || s.config.WorkspacePath == "" {
		t.Errorf("expected WorkspacePath to be resolved to absolute path, got %q", s.config.WorkspacePath)
	}
	if !filepath.IsAbs(s.config.WorkspacePath) {
		t.Errorf("expected resolved WorkspacePath to be absolute path, got %q", s.config.WorkspacePath)
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

func TestScheduler_VerboseLogging(t *testing.T) {
	tempWorkspace := t.TempDir()
	goFile := filepath.Join(tempWorkspace, "main.go")
	goSource := `package main
	func main() {}`
	if err := os.WriteFile(goFile, []byte(goSource), 0644); err != nil {
		t.Fatalf("failed to write mock Go file: %v", err)
	}

	// Case 1: Non-verbose mode (default)
	{
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cfg := Config{
			WorkspacePath: tempWorkspace,
			Verbose:       false,
		}
		spy := testutil.NewMockGraphService()
		s := NewScheduler(cfg, spy)
		s.ingest(IngestTrigger{Type: TriggerPeriodicScan})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if strings.Contains(output, "EXTRACTED KNOWLEDGE GRAPH SNAPSHOT") {
			t.Error("expected output NOT to contain KNOWLEDGE GRAPH SNAPSHOT in non-verbose mode")
		}
		if strings.Contains(output, "Starting full workspace scan") {
			t.Error("expected output NOT to contain Starting full workspace scan in non-verbose mode")
		}
		if strings.Contains(output, "Scanning modified file") {
			t.Error("expected output NOT to contain Scanning modified file in non-verbose mode")
		}
	}

	// Case 2: Verbose mode
	{
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		cfg := Config{
			WorkspacePath: tempWorkspace,
			Verbose:       true,
		}
		spy := testutil.NewMockGraphService()
		s := NewScheduler(cfg, spy)
		s.ingest(IngestTrigger{Type: TriggerPeriodicScan})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "EXTRACTED KNOWLEDGE GRAPH SNAPSHOT") {
			t.Error("expected output to contain KNOWLEDGE GRAPH SNAPSHOT in verbose mode")
		}
		if !strings.Contains(output, "Starting full workspace scan") {
			t.Error("expected output to contain Starting full workspace scan in verbose mode")
		}
		if !strings.Contains(output, "Scanning modified file") {
			t.Error("expected output to contain Scanning modified file in verbose mode")
		}
	}
}

