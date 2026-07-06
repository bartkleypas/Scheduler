package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"scheduler/internal/testutil"
)

func TestIsDir(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(tempFile, []byte("hello"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	if !isDir(tempDir) {
		t.Errorf("expected isDir(%s) to be true", tempDir)
	}

	if isDir(tempFile) {
		t.Errorf("expected isDir(%s) to be false", tempFile)
	}

	if isDir(filepath.Join(tempDir, "does-not-exist")) {
		t.Errorf("expected isDir to be false for non-existent path")
	}
}

func TestWatcher_BasicFileChanged(t *testing.T) {
	tempWorkspace := t.TempDir()

	// Ensure there is at least one subfolder and a file in it
	subDir := filepath.Join(tempWorkspace, "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	cfg := Config{
		WorkspacePath: tempWorkspace,
		EnableWatcher: true,
		WatchDebounce: 50 * time.Millisecond,
	}

	// Create scheduler with a dummy service
	s := NewScheduler(cfg, testutil.NewMockGraphService())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Run watcher in background
	go s.runWatcher(ctx)

	// Wait a moment for watcher to initialize and walk directories
	time.Sleep(100 * time.Millisecond)

	// Write a file in the watched directory
	testFile := filepath.Join(subDir, "app.go")
	if err := os.WriteFile(testFile, []byte("package pkg"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// Verify we receive a trigger
	select {
	case trigger := <-s.triggers:
		if trigger.Type != TriggerFileChanged {
			t.Errorf("expected trigger type %s, got %s", TriggerFileChanged, trigger.Type)
		}
		expectedPath := testFile
		// Normalize paths to absolute before comparison since runWatcher uses filepath.Abs
		absExpected, _ := filepath.Abs(expectedPath)
		absTrigger, _ := filepath.Abs(trigger.Path)
		if absTrigger != absExpected {
			t.Errorf("expected trigger path %s, got %s", absExpected, absTrigger)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for file watcher trigger")
	}
}

func TestWatcher_DynamicSubdirectory(t *testing.T) {
	tempWorkspace := t.TempDir()

	cfg := Config{
		WorkspacePath: tempWorkspace,
		EnableWatcher: true,
		WatchDebounce: 50 * time.Millisecond,
	}

	s := NewScheduler(cfg, testutil.NewMockGraphService())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go s.runWatcher(ctx)

	time.Sleep(100 * time.Millisecond)

	// Dynamically create a folder
	newFolder := filepath.Join(tempWorkspace, "new_folder")
	if err := os.MkdirAll(newFolder, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Wait for fsnotify to register the directory create event
	time.Sleep(100 * time.Millisecond)

	// Write a file to the new folder
	nestedFile := filepath.Join(newFolder, "nested.go")
	if err := os.WriteFile(nestedFile, []byte("package new_folder"), 0644); err != nil {
		t.Fatalf("failed to write nested file: %v", err)
	}

	// Verify we receive a trigger for the file write inside the dynamic folder
	select {
	case trigger := <-s.triggers:
		if trigger.Type != TriggerFileChanged {
			t.Errorf("expected trigger type %s, got %s", TriggerFileChanged, trigger.Type)
		}
		absExpected, _ := filepath.Abs(nestedFile)
		absTrigger, _ := filepath.Abs(trigger.Path)
		if absTrigger != absExpected {
			t.Errorf("expected trigger path %s, got %s", absExpected, absTrigger)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for dynamic subdirectory file watcher trigger")
	}
}
