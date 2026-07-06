package engine

import (
	"context"
	"scheduler/pkg/parser"
)

// GraphService defines the business logic for interacting with the Knowledge Graph.
// It decouples the ingestion engine and API from the underlying storage implementation.
type GraphService interface {
	// SaveGraph persists a parsed graph into the store.
	SaveGraph(ctx context.Context, g *parser.Graph) error
	
	// GetGraph retrieves the entire graph from the store.
	GetGraph(ctx context.Context) (*parser.Graph, error)
	
	// DeleteNode removes a node and its associated edges.
	DeleteNode(ctx context.Context, id string) error

	// DeleteFileGraph removes a File node and all nodes defined in it.
	DeleteFileGraph(ctx context.Context, relPath string) error

	// GetFileHash retrieves the cached hash for a file.
	GetFileHash(ctx context.Context, relPath string) (string, error)

	// SetFileHash inserts or updates the cached hash for a file.
	SetFileHash(ctx context.Context, relPath, hash string) error

	// DeleteFileHash deletes the cached hash for a file.
	DeleteFileHash(ctx context.Context, relPath string) error
	
	// Close closes the underlying storage connection.
	Close() error
}
