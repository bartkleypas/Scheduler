package engine

import (
	"io/fs"
	"path/filepath"
	"scheduler/internal/ignore"
)

// WorkspaceFiles holds lists of Go and Markdown source files discovered in the workspace.
type WorkspaceFiles struct {
	GoFiles       []string
	MarkdownFiles []string
}

// WalkWorkspace recursively scans the root directory, honoring gitignore and default exclusions.
func WalkWorkspace(rootPath string) (WorkspaceFiles, error) {
	var files WorkspaceFiles
	matcher := ignore.NewMatcher(rootPath)

	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return files, err
	}

	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from workspace root
		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}

		// Skip walking ignored directories completely
		if d.IsDir() {
			if matcher.IsIgnored(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if individual file matches ignore list
		if matcher.IsIgnored(relPath) {
			return nil
		}

		// Group by extension
		ext := filepath.Ext(path)
		switch ext {
		case ".go":
			files.GoFiles = append(files.GoFiles, path)
		case ".md":
			files.MarkdownFiles = append(files.MarkdownFiles, path)
		}

		return nil
	})

	return files, err
}
