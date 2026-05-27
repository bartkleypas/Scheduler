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
	
	// Close closes the underlying storage connection.
	Close() error
}
