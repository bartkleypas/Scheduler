package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkWorkspace(t *testing.T) {
	// Create temporary workspace
	tempRoot, err := os.MkdirTemp("", "workspace_walk_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempRoot)

	// Create directories
	dirs := []string{
		"pkg/engine",
		"cmd/scheduler",
		"node_modules/lodash",
		".git",
		"decisions",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(tempRoot, d), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", d, err)
		}
	}

	// Create files
	files := map[string]string{
		"go.mod":                                 "module test",
		"pkg/engine/scheduler.go":                "package engine",
		"cmd/scheduler/main.go":                  "package main",
		"node_modules/lodash/index.js":           "console.log()",
		".git/config":                            "[core]",
		"decisions/001-graph-storage.md":         "# Graph Storage",
		"README.md":                              "# Test Project",
		".gitignore":                             "node_modules/\n",
	}
	for relPath, content := range files {
		absPath := filepath.Join(tempRoot, relPath)
		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write file %s: %v", relPath, err)
		}
	}

	// Run Walker
	workspaceFiles, err := WalkWorkspace(tempRoot)
	if err != nil {
		t.Fatalf("WalkWorkspace failed: %v", err)
	}

	// Verify GoFiles contain correct items and skip git/node_modules
	goFilesMap := make(map[string]bool)
	for _, f := range workspaceFiles.GoFiles {
		rel, _ := filepath.Rel(tempRoot, f)
		goFilesMap[rel] = true
	}

	expectedGo := []string{
		"pkg/engine/scheduler.go",
		"cmd/scheduler/main.go",
	}
	for _, f := range expectedGo {
		if !goFilesMap[f] {
			t.Errorf("expected Go file %s was not found in walked list", f)
		}
	}

	ignoredGo := []string{
		"node_modules/lodash/index.js",
	}
	for _, f := range ignoredGo {
		if goFilesMap[f] {
			t.Errorf("ignored Go file %s was unexpectedly walked", f)
		}
	}

	// Verify MarkdownFiles contain correct items
	mdFilesMap := make(map[string]bool)
	for _, f := range workspaceFiles.MarkdownFiles {
		rel, _ := filepath.Rel(tempRoot, f)
		mdFilesMap[rel] = true
	}

	expectedMd := []string{
		"decisions/001-graph-storage.md",
		"README.md",
	}
	for _, f := range expectedMd {
		if !mdFilesMap[f] {
			t.Errorf("expected Markdown file %s was not found in walked list", f)
		}
	}
}
