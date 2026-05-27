package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreMatcher(t *testing.T) {
	// Create a temporary directory for testing .gitignore parsing
	tempDir, err := os.MkdirTemp("", "ignore_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write a mock .gitignore
	gitIgnoreContent := `
# A comment in gitignore
node_modules/
dist
*.log
`
	err = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitIgnoreContent), 0644)
	if err != nil {
		t.Fatalf("failed to write mock gitignore: %v", err)
	}

	matcher := NewMatcher(tempDir)

	tests := []struct {
		path    string
		ignored bool
	}{
		{".git/config", true},
		{".knowledge/graph.db", true},
		{"node_modules/lodash/index.js", true},
		{"web/node_modules/react.js", true},
		{"dist/bundle.js", true},
		{"app.log", true},
		{"src/main.go", false},
		{"pkg/engine/scheduler.go", false},
	}

	for _, tt := range tests {
		result := matcher.IsIgnored(tt.path)
		if result != tt.ignored {
			t.Errorf("IsIgnored(%q) = %v; want %v", tt.path, result, tt.ignored)
		}
	}
}
