package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

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
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS nodes (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		properties TEXT
	);
	CREATE TABLE IF NOT EXISTS edges (
		from TEXT NOT NULL,
		to TEXT NOT NULL,
		type TEXT NOT NULL,
		PRIMARY KEY (from, to, type),
		FOREIGN KEY (from) REFERENCES nodes(id) ON DELETE CASCADE,
		FOREIGN KEY (to) REFERENCES nodes(id) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);
	`
	_, err := s.db.Exec(schema)
	return err
}

// UpsertNode inserts or updates a node.
func (s *SQLiteStore) UpsertNode(ctx context.Context, node Node) error {
	props, err := json.Marshal(node.Properties)
	if err != nil {
		return err
	}

	query := `INSERT INTO nodes (id, type, properties) VALUES (?, ?, ?) 
	          ON CONFLICT(id) DO UPDATE SET type=excluded.type, properties=excluded.properties`
	_, err = s.db.ExecContext(ctx, query, node.ID, node.Type, string(props))
	return err
}

// UpsertEdge inserts or updates an edge.
func (s *SQLiteStore) UpsertEdge(ctx context.Context, edge Edge) error {
	query := `INSERT INTO edges (from, to, type) VALUES (?, ?, ?) 
	          ON CONFLICT(from, to, type) DO NOTHING`
	_, err := s.db.ExecContext(ctx, query, edge.From, edge.To, edge.Type)
	return err
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
