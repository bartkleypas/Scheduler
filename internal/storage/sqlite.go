package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// Node represents a graph node in storage.
type Node struct {
	ID         string
	Type       string
	Properties json.RawMessage
}

// Edge represents a graph edge in storage.
type Edge struct {
	From string
	To   string
	Type string
}

// SQLiteStore handles the persistence of the Knowledge Graph.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore initializes a new SQLite store and runs migrations.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	// Ensure parent directory exists before opening database
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory %s: %w", dir, err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	// Enable foreign key constraints for cascading deletes
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys constraint: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	// 1. Create schema_version table
	initSchema := `
	CREATE TABLE IF NOT EXISTS schema_version (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := s.db.Exec(initSchema); err != nil {
		return fmt.Errorf("failed to initialize schema_version: %w", err)
	}

	// 2. Define sequential migrations
	var migrations = []struct {
		version int
		query   string
	}{
		{
			version: 1,
			query: `
				CREATE TABLE IF NOT EXISTS nodes (
					id TEXT PRIMARY KEY,
					type TEXT NOT NULL,
					properties TEXT
				);
				CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);
			`,
		},
		{
			version: 2,
			query: `
				CREATE TABLE IF NOT EXISTS edges (
					"from" TEXT NOT NULL,
					"to" TEXT NOT NULL,
					type TEXT NOT NULL,
					PRIMARY KEY ("from", "to", type),
					FOREIGN KEY ("from") REFERENCES nodes(id) ON DELETE CASCADE,
					FOREIGN KEY ("to") REFERENCES nodes(id) ON DELETE CASCADE
				);
				CREATE INDEX IF NOT EXISTS idx_edges_from ON edges("from");
				CREATE INDEX IF NOT EXISTS idx_edges_to ON edges("to");
			`,
		},
	}

	// 3. Apply migrations sequentially
	for _, m := range migrations {
		var count int
		err := s.db.QueryRow("SELECT COUNT(*) FROM schema_version WHERE version = ?", m.version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration %d: %w", m.version, err)
		}

		if count == 0 {
			// Apply migration in a transaction
			tx, err := s.db.Begin()
			if err != nil {
				return err
			}

			if _, err := tx.Exec(m.query); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %d: %w", m.version, err)
			}

			if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (?)", m.version); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to log migration %d: %w", m.version, err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration %d: %w", m.version, err)
			}
			fmt.Printf("[Migrations] Applied migration version %d successfully.\n", m.version)
		}
	}

	return nil
}

// UpsertNode inserts or updates a node.
func (s *SQLiteStore) UpsertNode(ctx context.Context, node Node) error {
	// Use the RawMessage directly as a string to avoid double-encoding
	query := `INSERT INTO nodes (id, type, properties) VALUES (?, ?, ?) 
	          ON CONFLICT(id) DO UPDATE SET type=excluded.type, properties=excluded.properties`
	_, err := s.db.ExecContext(ctx, query, node.ID, node.Type, string(node.Properties))
	return err
}

// UpsertEdge inserts or updates an edge.
func (s *SQLiteStore) UpsertEdge(ctx context.Context, edge Edge) error {
	query := `INSERT INTO edges ("from", "to", type) VALUES (?, ?, ?) 
	          ON CONFLICT("from", "to", type) DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, edge.From, edge.To, edge.Type)
	return err
}

// GetNodes retrieves all stored nodes.
func (s *SQLiteStore) GetNodes(ctx context.Context) ([]Node, error) {
	query := `SELECT id, type, properties FROM nodes`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query nodes: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		var propsStr sql.NullString
		if err := rows.Scan(&n.ID, &n.Type, &propsStr); err != nil {
			return nil, fmt.Errorf("failed to scan node: %w", err)
		}
		if propsStr.Valid {
			n.Properties = json.RawMessage(propsStr.String)
		}
		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("nodes row error: %w", err)
	}
	return nodes, nil
}

// GetEdges retrieves all stored edges.
func (s *SQLiteStore) GetEdges(ctx context.Context) ([]Edge, error) {
	// from and to are SQLite keywords, so we quote them
	query := `SELECT "from", "to", type FROM edges`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges: %w", err)
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.From, &e.To, &e.Type); err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}
		edges = append(edges, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("edges row error: %w", err)
	}
	return edges, nil
}

// DeleteNode removes a node by its ID. SQLite foreign keys cascade constraints
// will automatically delete any associated incoming or outgoing edges.
func (s *SQLiteStore) DeleteNode(ctx context.Context, id string) error {
	query := `DELETE FROM nodes WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete node %s: %w", id, err)
	}
	return nil
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// GenerateUUIDv7 creates a new UUIDv7.
func GenerateUUIDv7() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.NewString() // Fallback to v4 if v7 fails
	}
	return id.String()
}
