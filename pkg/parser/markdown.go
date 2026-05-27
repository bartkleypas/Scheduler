package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ParseMarkdownDebate extracts graph nodes and edges from a structured markdown debate file.
func ParseMarkdownDebate(absPath string, rootPath string) (*Graph, error) {
	file, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	relPath, err := filepath.Rel(rootPath, absPath)
	if err != nil {
		relPath = filepath.Base(absPath)
	}

	g := NewGraph()
	scanner := bufio.NewScanner(file)

	// Frontmatter parsing state
	inFrontmatter := false
	frontmatterChecked := false
	frontmatterLines := []string{}
	markdownBodyLines := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !frontmatterChecked {
			if trimmed == "---" {
				if !inFrontmatter {
					inFrontmatter = true
				} else {
					inFrontmatter = false
					frontmatterChecked = true
				}
				continue
			}
			if inFrontmatter {
				frontmatterLines = append(frontmatterLines, line)
				continue
			}
		}

		markdownBodyLines = append(markdownBodyLines, line)
	}

	// Parse YAML key-values
	properties := make(map[string]interface{})
	var targets []string
	var debateID string
	var debateTitle string
	var debateStatus string

	// Simple YAML Frontmatter parser
	isParsingList := false
	listKey := ""
	for _, fLine := range frontmatterLines {
		trimmed := strings.TrimSpace(fLine)
		if trimmed == "" {
			continue
		}

		// Handle list items (e.g. - targets)
		if strings.HasPrefix(trimmed, "-") && isParsingList {
			item := strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
			item = strings.Trim(item, `"'`)
			if listKey == "targets" {
				targets = append(targets, item)
			}
			continue
		}

		parts := strings.SplitN(fLine, ":", 2)
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)

		// Reset list state
		isParsingList = false

		if val == "" {
			// This might be the start of a YAML list
			isParsingList = true
			listKey = key
			continue
		}

		// Inline array parsing (e.g. [target1, target2])
		if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
			arrayContent := strings.Trim(val, "[]")
			items := strings.Split(arrayContent, ",")
			for _, item := range items {
				cItem := strings.TrimSpace(item)
				cItem = strings.Trim(cItem, `"'`)
				if cItem != "" {
					if key == "targets" {
						targets = append(targets, cItem)
					}
				}
			}
			continue
		}

		switch key {
		case "id":
			debateID = val
		case "title":
			debateTitle = val
		case "status":
			debateStatus = val
		}
		properties[key] = val
	}

	// Fallback to filename if ID not provided
	if debateID == "" {
		base := filepath.Base(absPath)
		debateID = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// Fallback to first H1 header in markdown body if Title not in frontmatter
	if debateTitle == "" {
		for _, line := range markdownBodyLines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "# ") {
				debateTitle = strings.TrimPrefix(trimmed, "# ")
				break
			}
		}
		if debateTitle == "" {
			debateTitle = debateID
		}
	}
	properties["title"] = debateTitle
	properties["filepath"] = relPath

	// Map Decision or Debate Node
	nodeType := "Debate"
	if strings.ToLower(debateStatus) == "resolved" {
		nodeType = "Decision"
	}

	g.AddNode(Node{
		ID:         debateID,
		Type:       nodeType,
		Properties: properties,
	})

	// Add the Debate file node itself
	g.AddNode(Node{
		ID:   relPath,
		Type: "File",
		Properties: map[string]interface{}{
			"filepath": relPath,
		},
	})

	// Link Debate to its file
	g.AddEdge(debateID, relPath, "DEFINED_IN")

	// Parse Targets
	for _, target := range targets {
		// Ensure target paths are consistent relative paths
		cleanTarget := filepath.ToSlash(filepath.Clean(target))

		// Ensure the target File exists in the graph as a node placeholder
		g.AddNode(Node{
			ID:   cleanTarget,
			Type: "File",
			Properties: map[string]interface{}{
				"filepath": cleanTarget,
			},
		})

		// Create edge (DebateNode) -[TARGETS]-> (TargetFileNode)
		g.AddEdge(debateID, cleanTarget, "TARGETS")
	}

	// Parse tasks/todo items from body
	for _, line := range markdownBodyLines {
		trimmed := strings.TrimSpace(line)
		var isTodo, isDone bool
		var taskDesc string

		if strings.HasPrefix(trimmed, "- [ ]") {
			isTodo = true
			isDone = false
			taskDesc = strings.TrimSpace(strings.TrimPrefix(trimmed, "- [ ]"))
		} else if strings.HasPrefix(trimmed, "- [x]") {
			isTodo = true
			isDone = true
			taskDesc = strings.TrimSpace(strings.TrimPrefix(trimmed, "- [x]"))
		}

		if isTodo && taskDesc != "" {
			// Generate a task ID (DebateID + TaskDesc slug)
			taskSlug := strings.ToLower(strings.ReplaceAll(taskDesc, " ", "-"))
			taskID := debateID + "-task-" + taskSlug

			g.AddNode(Node{
				ID:   taskID,
				Type: "Task",
				Properties: map[string]interface{}{
					"description": taskDesc,
					"completed":   isDone,
				},
			})

			// Link task to the debate: (Task) -[RESOLVES]-> (Debate)
			g.AddEdge(taskID, debateID, "RESOLVES")
		}
	}

	return g, nil
}
