package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestNewSQLiteStore(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "sub", "test.db") // nested folder to test directory creation

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("database file was not created at %s", dbPath)
	}

	// Verify migration tables exist by executing a query on schema_version
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count)
	if err != nil {
		t.Errorf("failed to query schema_version: %v", err)
	}

	// Double check tables exist
	tablesQuery := "SELECT name FROM sqlite_master WHERE type='table' AND name IN ('nodes', 'edges')"
	rows, err := store.db.Query(tablesQuery)
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer rows.Close()

	foundTables := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("failed to scan table name: %v", err)
		}
		foundTables[name] = true
	}
	if !foundTables["nodes"] {
		t.Error("nodes table was not created")
	}
	if !foundTables["edges"] {
		t.Error("edges table was not created")
	}
}

func TestSQLiteStore_NodeOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	node1 := Node{
		ID:         "node-1",
		Type:       "File",
		Properties: json.RawMessage(`{"path": "pkg/engine/scheduler.go"}`),
	}

	// Test UpsertNode (Insert)
	if err := store.UpsertNode(ctx, node1); err != nil {
		t.Fatalf("failed to upsert node: %v", err)
	}

	// Test GetNodes
	nodes, err := store.GetNodes(ctx)
	if err != nil {
		t.Fatalf("failed to get nodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].ID != node1.ID || nodes[0].Type != node1.Type || string(nodes[0].Properties) != string(node1.Properties) {
		t.Errorf("node mismatch. expected %+v, got %+v", node1, nodes[0])
	}

	// Test UpsertNode (Update properties and type)
	node1Updated := Node{
		ID:         "node-1",
		Type:       "UpdatedFile",
		Properties: json.RawMessage(`{"path": "pkg/engine/scheduler_updated.go"}`),
	}
	if err := store.UpsertNode(ctx, node1Updated); err != nil {
		t.Fatalf("failed to update node: %v", err)
	}

	nodes, err = store.GetNodes(ctx)
	if err != nil {
		t.Fatalf("failed to get nodes: %v", err)
	}
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Type != "UpdatedFile" || string(nodes[0].Properties) != string(node1Updated.Properties) {
		t.Errorf("node properties update failed. got %+v", nodes[0])
	}
}

func TestSQLiteStore_EdgeOperations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Insert nodes first because of foreign key constraints
	node1 := Node{ID: "node-1", Type: "File", Properties: json.RawMessage(`{}`)}
	node2 := Node{ID: "node-2", Type: "File", Properties: json.RawMessage(`{}`)}
	if err := store.UpsertNode(ctx, node1); err != nil {
		t.Fatalf("failed to insert node 1: %v", err)
	}
	if err := store.UpsertNode(ctx, node2); err != nil {
		t.Fatalf("failed to insert node 2: %v", err)
	}

	edge := Edge{
		From: "node-1",
		To:   "node-2",
		Type: "IMPORTS",
	}

	// Test UpsertEdge (Insert)
	if err := store.UpsertEdge(ctx, edge); err != nil {
		t.Fatalf("failed to upsert edge: %v", err)
	}

	// Test UpsertEdge duplicate DO NOTHING
	if err := store.UpsertEdge(ctx, edge); err != nil {
		t.Fatalf("failed duplicate upsert edge: %v", err)
	}

	// Test GetEdges
	edges, err := store.GetEdges(ctx)
	if err != nil {
		t.Fatalf("failed to get edges: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
	if edges[0] != edge {
		t.Errorf("edge mismatch. expected %+v, got %+v", edge, edges[0])
	}
}

func TestSQLiteStore_CascadeDelete(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	store, err := NewSQLiteStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Setup Nodes
	node1 := Node{ID: "node-1", Type: "File", Properties: json.RawMessage(`{}`)}
	node2 := Node{ID: "node-2", Type: "File", Properties: json.RawMessage(`{}`)}
	if err := store.UpsertNode(ctx, node1); err != nil {
		t.Fatalf("failed to insert node 1: %v", err)
	}
	if err := store.UpsertNode(ctx, node2); err != nil {
		t.Fatalf("failed to insert node 2: %v", err)
	}

	// Setup Edge
	edge := Edge{From: "node-1", To: "node-2", Type: "IMPORTS"}
	if err := store.UpsertEdge(ctx, edge); err != nil {
		t.Fatalf("failed to upsert edge: %v", err)
	}

	// Test DeleteNode
	if err := store.DeleteNode(ctx, "node-1"); err != nil {
		t.Fatalf("failed to delete node: %v", err)
	}

	// Verify node deleted
	nodes, err := store.GetNodes(ctx)
	if err != nil {
		t.Fatalf("failed to get nodes: %v", err)
	}
	for _, n := range nodes {
		if n.ID == "node-1" {
			t.Error("node-1 was not deleted from nodes table")
		}
	}

	// Verify edge deleted via CASCADE
	edges, err := store.GetEdges(ctx)
	if err != nil {
		t.Fatalf("failed to get edges: %v", err)
	}
	if len(edges) != 0 {
		t.Errorf("expected 0 edges after cascade delete, got %d: %+v", len(edges), edges)
	}
}

func TestGenerateUUIDv7(t *testing.T) {
	idStr := GenerateUUIDv7()

	// Verify it's a valid UUID
	parsed, err := uuid.Parse(idStr)
	if err != nil {
		t.Fatalf("GenerateUUIDv7 generated invalid UUID format: %v", err)
	}

	// Verify it is UUID version 7
	if parsed.Version() != 7 {
		t.Errorf("expected UUID version 7, got %d", parsed.Version())
	}
}
