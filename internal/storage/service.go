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
	nodes, err := s.store.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	edges, err := s.store.GetEdges(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges: %w", err)
	}

	g := parser.NewGraph()

	for _, rawNode := range nodes {
		var properties map[string]interface{}
		if len(rawNode.Properties) > 0 {
			if err := json.Unmarshal(rawNode.Properties, &properties); err != nil {
				return nil, fmt.Errorf("failed to unmarshal properties for node %s: %w", rawNode.ID, err)
			}
		}
		g.AddNode(parser.Node{
			ID:         rawNode.ID,
			Type:       rawNode.Type,
			Properties: properties,
		})
	}

	for _, rawEdge := range edges {
		g.Edges = append(g.Edges, parser.Edge{
			From: rawEdge.From,
			To:   rawEdge.To,
			Type: rawEdge.Type,
		})
	}

	return g, nil
}

// DeleteNode removes a node and its cascading edges.
func (s *SQLiteGraphService) DeleteNode(ctx context.Context, id string) error {
	return s.store.DeleteNode(ctx, id)
}

// Close closes the database connection.
func (s *SQLiteGraphService) Close() error {
	return s.store.Close()
}
