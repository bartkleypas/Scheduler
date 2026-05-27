package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseMarkdownDebate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "markdown_parser_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockFileContent := `---
id: graph-storage
title: Graph Storage Engine
status: Resolved
decision: Option A (SQLite + JSON)
targets:
  - pkg/engine/db.go
  - pkg/engine/schema.sql
---

# Graph Storage Engine

SQLite was chosen as the relational local storage layer for the knowledge graph.

- [ ] Implement SQLite schema initialization
- [x] Select SQLite library driver
`

	mockFilePath := filepath.Join(tempDir, "001-graph-storage.md")
	err = os.WriteFile(mockFilePath, []byte(mockFileContent), 0644)
	if err != nil {
		t.Fatalf("failed to write mock markdown file: %v", err)
	}

	graph, err := ParseMarkdownDebate(mockFilePath, tempDir)
	if err != nil {
		t.Fatalf("ParseMarkdownDebate failed: %v", err)
	}

	// 1. Verify the Debate/Decision Node is created
	decisionNode, ok := graph.Nodes["graph-storage"]
	if !ok {
		t.Fatal("expected decision node 'graph-storage' was not found in graph")
	}
	if decisionNode.Type != "Decision" {
		t.Errorf("expected decision node type 'Decision'; got %q", decisionNode.Type)
	}
	if decisionNode.Properties["title"] != "Graph Storage Engine" {
		t.Errorf("expected decision title 'Graph Storage Engine'; got %q", decisionNode.Properties["title"])
	}
	if decisionNode.Properties["status"] != "Resolved" {
		t.Errorf("expected decision status 'Resolved'; got %q", decisionNode.Properties["status"])
	}

	// 2. Verify targets are parsed and File nodes created
	expectedTargets := []string{
		"pkg/engine/db.go",
		"pkg/engine/schema.sql",
	}
	for _, target := range expectedTargets {
		fileNode, ok := graph.Nodes[target]
		if !ok {
			t.Errorf("expected target file node %q was not found", target)
		}
		if fileNode.Type != "File" {
			t.Errorf("expected target node type 'File'; got %q", fileNode.Type)
		}
	}

	// Verify target edges exist
	targetEdgesCount := 0
	for _, edge := range graph.Edges {
		if edge.From == "graph-storage" && edge.Type == "TARGETS" {
			targetEdgesCount++
		}
	}
	if targetEdgesCount != 2 {
		t.Errorf("expected 2 TARGETS edges; got %d", targetEdgesCount)
	}

	// 3. Verify tasks are parsed
	todoTaskID := "graph-storage-task-implement-sqlite-schema-initialization"
	todoTask, ok := graph.Nodes[todoTaskID]
	if !ok {
		t.Errorf("expected Task node %q not found", todoTaskID)
	}
	if todoTask.Properties["completed"] != false {
		t.Errorf("expected todo task completed=false; got %v", todoTask.Properties["completed"])
	}

	doneTaskID := "graph-storage-task-select-sqlite-library-driver"
	doneTask, ok := graph.Nodes[doneTaskID]
	if !ok {
		t.Errorf("expected Task node %q not found", doneTaskID)
	}
	if doneTask.Properties["completed"] != true {
		t.Errorf("expected done task completed=true; got %v", doneTask.Properties["completed"])
	}

	// Verify resolves edge
	hasResolvesEdge := false
	for _, edge := range graph.Edges {
		if edge.From == todoTaskID && edge.To == "graph-storage" && edge.Type == "RESOLVES" {
			hasResolvesEdge = true
		}
	}
	if !hasResolvesEdge {
		t.Error("expected RESOLVES edge connecting Task to Decision was not found")
	}
}
