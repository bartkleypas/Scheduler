package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"scheduler/pkg/parser"
)

// SQLiteGraphService implements engine.GraphService using a SQLite backend.
type SQLiteGraphService struct {
	store *SQLiteStore
}

// NewSQLiteGraphService creates a new instance of the GraphService.
func NewSQLiteGraphService(store *SQLiteStore) *SQLiteGraphService {
	return &SQLiteGraphService{store: store}
}

// SaveGraph takes a parsed graph and persists it into SQLite.
func (s *SQLiteGraphService) SaveGraph(ctx context.Context, g *parser.Graph) error {
	if g == nil {
		return nil
	}

	// 1. Save Nodes
	for _, node := range g.Nodes {
		nodeProps, err := json.Marshal(node.Properties)
		if err != nil {
			return fmt.Errorf("failed to marshal properties for node %s: %w", node.ID, err)
		}

		err = s.store.UpsertNode(ctx, Node{
			ID:         node.ID,
			Type:       node.Type,
			Properties: nodeProps,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert node %s: %w", node.ID, err)
		}
	}

	// 2. Save Edges
	for _, edge := range g.Edges {
		err := s.store.UpsertEdge(ctx, Edge{
			From: edge.From,
			To:   edge.To,
			Type: edge.Type,
		})
		if err != nil {
			return fmt.Errorf("failed to upsert edge %s->%s: %w", edge.From, edge.To, err)
		}
	}

	return nil
}

// GetGraph retrieves the current state of the graph from storage.
func (s *SQLiteGraphService) GetGraph(ctx context.Context) (*parser.Graph, error) {
	// Note: In a real production system, we might want pagination or filtering.
	// For the scheduler knowledge graph, we currently return the full set for the D3 visualization.
	
	// This is a simplified implementation of retrieval.
	// We'll need to add helper methods to SQLiteStore to fetch all nodes and edges.
	return nil, fmt.Errorf("GetGraph not yet implemented in SQLiteGraphService")
}

// DeleteNode removes a node and its cascading edges.
func (s *SQLiteGraphService) DeleteNode(ctx context.Context, id string) error {
	// We can add this to SQLiteStore later.
	return fmt.Errorf("DeleteNode not yet implemented in SQLiteGraphService")
}

// Close closes the database connection.
func (s *SQLiteGraphService) Close() error {
	return s.store.Close()
}
