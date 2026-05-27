package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"scheduler/pkg/parser"
)

func TestSchedulerIntegration(t *testing.T) {
	// 1. Create a mock workspace
	tempWorkspace, err := os.MkdirTemp("", "scheduler_integration_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempWorkspace)

	// Create directories
	dirs := []string{
		"pkg/engine",
		"decisions",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(tempWorkspace, d), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	// Create a mock Go file
	goSource := `package main
import "fmt"
type App struct {}
func main() {
	fmt.Println("Integration Test")
}
`
	err = os.WriteFile(filepath.Join(tempWorkspace, "pkg/engine/app.go"), []byte(goSource), 0644)
	if err != nil {
		t.Fatalf("failed to write mock go file: %v", err)
	}

	// Create a mock Debate file
	debateContent := `---
id: db-choice
title: Database Choice
status: Resolved
decision: SQLite
targets:
  - pkg/engine/app.go
---
# Database Choice
- [ ] Initialize db connection
`
	err = os.WriteFile(filepath.Join(tempWorkspace, "decisions/001-db-choice.md"), []byte(debateContent), 0644)
	if err != nil {
		t.Fatalf("failed to write mock debate file: %v", err)
	}

	// 2. Initialize Scheduler
	cfg := Config{
		WorkspacePath: tempWorkspace,
		DBPath:        filepath.Join(tempWorkspace, ".knowledge/graph.db"),
		ScanInterval:  0,     // Disable ticker for deterministic test
		EnableWatcher: false, // Disable watcher
	}

	scheduler := NewScheduler(cfg)

	// Capture trigger channel in testing context to verify it works
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := scheduler.Start(ctx); err != nil {
		t.Fatalf("failed to start scheduler: %v", err)
	}

	// 3. Test Full Walk and Parse manually
	workspaceFiles, err := WalkWorkspace(tempWorkspace)
	if err != nil {
		t.Fatalf("WalkWorkspace failed: %v", err)
	}

	if len(workspaceFiles.GoFiles) != 1 {
		t.Errorf("expected 1 Go file; got %d", len(workspaceFiles.GoFiles))
	}
	if len(workspaceFiles.MarkdownFiles) != 1 {
		t.Errorf("expected 1 Markdown file; got %d", len(workspaceFiles.MarkdownFiles))
	}

	// Parse Go source
	goGraph, err := parser.ParseGoSource(workspaceFiles.GoFiles[0], tempWorkspace)
	if err != nil {
		t.Fatalf("ParseGoSource failed: %v", err)
	}

	// Parse Markdown debate
	mdGraph, err := parser.ParseMarkdownDebate(workspaceFiles.MarkdownFiles[0], tempWorkspace)
	if err != nil {
		t.Fatalf("ParseMarkdownDebate failed: %v", err)
	}

	// Combine
	fullGraph := parser.NewGraph()
	fullGraph.Merge(goGraph)
	fullGraph.Merge(mdGraph)

	// 4. Verify Nodes and Relationships
	// File nodes
	if _, ok := fullGraph.Nodes["pkg/engine/app.go"]; !ok {
		t.Error("expected File node 'pkg/engine/app.go' not found")
	}

	// Go AST Symbols
	if _, ok := fullGraph.Nodes["main.App"]; !ok {
		t.Error("expected Symbol node 'main.App' not found")
	}

	// Markdown Decisions
	if _, ok := fullGraph.Nodes["db-choice"]; !ok {
		t.Error("expected Decision node 'db-choice' not found")
	}

	// Tasks
	taskID := "db-choice-task-initialize-db-connection"
	if _, ok := fullGraph.Nodes[taskID]; !ok {
		t.Error("expected Task node not found")
	}

	// TARGETS relationship
	hasTargetsRelation := false
	for _, edge := range fullGraph.Edges {
		if edge.From == "db-choice" && edge.To == "pkg/engine/app.go" && edge.Type == "TARGETS" {
			hasTargetsRelation = true
		}
	}
	if !hasTargetsRelation {
		t.Error("expected TARGETS edge between 'db-choice' and 'pkg/engine/app.go' not found")
	}

	// Let's run trigger channel manually to check event loop
	scheduler.SyncNow()
	
	// Wait a tiny bit for the goroutines to digest the manual sync trigger
	time.Sleep(50 * time.Millisecond)
}
