package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Matcher evaluates if a file path matches gitignore patterns.
type Matcher struct {
	patterns []string
}

// NewMatcher reads a .gitignore file and instantiates a pattern matcher.
func NewMatcher(rootPath string) *Matcher {
	m := &Matcher{
		patterns: []string{
			".git",
			".knowledge",
		},
	}

	gitIgnorePath := filepath.Join(rootPath, ".gitignore")
	file, err := os.Open(gitIgnorePath)
	if err != nil {
		return m // No .gitignore found, return defaults
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		m.patterns = append(m.patterns, line)
	}

	return m
}

// IsIgnored evaluates if a relative path matches any ignore patterns.
func (m *Matcher) IsIgnored(relPath string) bool {
	if relPath == "" || relPath == "." {
		return false
	}

	// Standardize separators to forward slashes for matching
	cleanPath := filepath.ToSlash(relPath)
	segments := strings.Split(cleanPath, "/")

	for _, pattern := range m.patterns {
		pat := filepath.ToSlash(pattern)
		pat = strings.TrimSuffix(strings.TrimPrefix(pat, "/"), "/")

		// 1. Direct segment match (e.g., "node_modules" matches "web/node_modules/lodash")
		for _, seg := range segments {
			if seg == pat {
				return true
			}
		}

		// 2. Prefix match (e.g., "dist/" matches "dist/bundle.js")
		if strings.HasPrefix(cleanPath, pat+"/") {
			return true
		}

		// 3. Simple wildcard match using filepath.Match
		matched, _ := filepath.Match(pat, cleanPath)
		if matched {
			return true
		}

		// 4. Match against individual segment wildcards
		for _, seg := range segments {
			matched, _ := filepath.Match(pat, seg)
			if matched {
				return true
			}
		}
	}

	return false
}
