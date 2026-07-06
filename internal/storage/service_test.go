package storage

import (
	"context"
	"path/filepath"
	"testing"

	"scheduler/pkg/parser"
)

func TestSQLiteGraphService_SaveAndGetGraph(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	service := NewSQLiteGraphService(store)
	ctx := context.Background()

	// 1. Save empty graph (should not error and do nothing)
	if err := service.SaveGraph(ctx, nil); err != nil {
		t.Fatalf("SaveGraph(nil) returned error: %v", err)
	}

	// 2. Create a test graph
	g := parser.NewGraph()
	g.AddNode(parser.Node{
		ID:   "file-1.go",
		Type: "File",
		Properties: map[string]interface{}{
			"path": "file-1.go",
			"size": float64(1234), // JSON unmarshal converts numbers to float64 by default
		},
	})
	g.AddNode(parser.Node{
		ID:   "main.App",
		Type: "Symbol",
		Properties: map[string]interface{}{
			"kind": "struct",
		},
	})
	g.Edges = append(g.Edges, parser.Edge{
		From: "main.App",
		To:   "file-1.go",
		Type: "DEFINED_IN",
	})

	// 3. Save graph
	if err := service.SaveGraph(ctx, g); err != nil {
		t.Fatalf("SaveGraph failed: %v", err)
	}

	// 4. Retrieve graph
	retrieved, err := service.GetGraph(ctx)
	if err != nil {
		t.Fatalf("GetGraph failed: %v", err)
	}

	// 5. Assert nodes exist and properties match
	if len(retrieved.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(retrieved.Nodes))
	}

	node1, ok := retrieved.Nodes["file-1.go"]
	if !ok {
		t.Fatalf("file-1.go not found in retrieved graph")
	}
	if node1.Type != "File" {
		t.Errorf("expected node type 'File', got '%s'", node1.Type)
	}
	if node1.Properties["path"] != "file-1.go" {
		t.Errorf("expected property 'path'='file-1.go', got '%v'", node1.Properties["path"])
	}
	if node1.Properties["size"] != float64(1234) {
		t.Errorf("expected property 'size'=1234, got '%v'", node1.Properties["size"])
	}

	node2, ok := retrieved.Nodes["main.App"]
	if !ok {
		t.Fatalf("main.App not found in retrieved graph")
	}
	if node2.Properties["kind"] != "struct" {
		t.Errorf("expected property 'kind'='struct', got '%v'", node2.Properties["kind"])
	}

	// 6. Assert edges match
	if len(retrieved.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(retrieved.Edges))
	}
	edge := retrieved.Edges[0]
	if edge.From != "main.App" || edge.To != "file-1.go" || edge.Type != "DEFINED_IN" {
		t.Errorf("edge mismatch. expected main.App -[DEFINED_IN]-> file-1.go, got %+v", edge)
	}
}

func TestSQLiteGraphService_DeleteNode(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	service := NewSQLiteGraphService(store)
	ctx := context.Background()

	g := parser.NewGraph()
	g.AddNode(parser.Node{ID: "node-1", Type: "File"})
	g.AddNode(parser.Node{ID: "node-2", Type: "File"})
	g.Edges = append(g.Edges, parser.Edge{From: "node-1", To: "node-2", Type: "DEPENDS"})

	if err := service.SaveGraph(ctx, g); err != nil {
		t.Fatalf("SaveGraph failed: %v", err)
	}

	// Delete node-1
	if err := service.DeleteNode(ctx, "node-1"); err != nil {
		t.Fatalf("DeleteNode failed: %v", err)
	}

	// Verify only node-2 remains, and 0 edges remain (due to cascade delete)
	retrieved, err := service.GetGraph(ctx)
	if err != nil {
		t.Fatalf("GetGraph failed: %v", err)
	}

	if len(retrieved.Nodes) != 1 {
		t.Fatalf("expected 1 node remaining, got %d", len(retrieved.Nodes))
	}
	if _, ok := retrieved.Nodes["node-2"]; !ok {
		t.Error("expected node-2 to remain in the graph")
	}
	if len(retrieved.Edges) != 0 {
		t.Errorf("expected 0 edges remaining, got %d", len(retrieved.Edges))
	}
}

func TestSQLiteGraphService_FileHashing(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	service := NewSQLiteGraphService(store)
	ctx := context.Background()

	// 1. Get non-existent hash (should return empty string, no error)
	hash, err := service.GetFileHash(ctx, "nonexistent.go")
	if err != nil {
		t.Fatalf("GetFileHash failed: %v", err)
	}
	if hash != "" {
		t.Errorf("expected empty hash, got %q", hash)
	}

	// 2. Set hash
	err = service.SetFileHash(ctx, "app.go", "hash123")
	if err != nil {
		t.Fatalf("SetFileHash failed: %v", err)
	}

	// 3. Get hash and verify
	hash, err = service.GetFileHash(ctx, "app.go")
	if err != nil {
		t.Fatalf("GetFileHash failed: %v", err)
	}
	if hash != "hash123" {
		t.Errorf("expected hash123, got %q", hash)
	}

	// 4. Update hash
	err = service.SetFileHash(ctx, "app.go", "hash456")
	if err != nil {
		t.Fatalf("SetFileHash update failed: %v", err)
	}

	// Verify updated
	hash, err = service.GetFileHash(ctx, "app.go")
	if err != nil {
		t.Fatalf("GetFileHash failed: %v", err)
	}
	if hash != "hash456" {
		t.Errorf("expected hash456, got %q", hash)
	}

	// 5. Delete hash
	err = service.DeleteFileHash(ctx, "app.go")
	if err != nil {
		t.Fatalf("DeleteFileHash failed: %v", err)
	}

	// Verify deleted
	hash, err = service.GetFileHash(ctx, "app.go")
	if err != nil {
		t.Fatalf("GetFileHash failed: %v", err)
	}
	if hash != "" {
		t.Errorf("expected empty hash after delete, got %q", hash)
	}
}
