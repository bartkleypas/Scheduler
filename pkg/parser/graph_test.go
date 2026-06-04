package parser

import "testing"

func TestGraph_Merge(t *testing.T) {
	// 1. Merge nil graph
	g1 := NewGraph()
	g1.Merge(nil)
	if len(g1.Nodes) != 0 || len(g1.Edges) != 0 {
		t.Errorf("expected empty graph after merging nil, got %d nodes and %d edges", len(g1.Nodes), len(g1.Edges))
	}

	// 2. Merge non-nil graph
	g1.AddNode(Node{ID: "node-1", Type: "File"})
	g1.AddEdge("node-1", "node-2", "IMPORTS")

	g2 := NewGraph()
	g2.AddNode(Node{ID: "node-2", Type: "File"})
	g2.AddNode(Node{ID: "node-3", Type: "Symbol"})
	g2.AddEdge("node-2", "node-3", "DEFINED_IN")

	g1.Merge(g2)

	// Check nodes merged
	if len(g1.Nodes) != 3 {
		t.Errorf("expected 3 nodes after merge, got %d", len(g1.Nodes))
	}
	for _, id := range []string{"node-1", "node-2", "node-3"} {
		if _, ok := g1.Nodes[id]; !ok {
			t.Errorf("missing node %s after merge", id)
		}
	}

	// Check edges merged
	if len(g1.Edges) != 2 {
		t.Errorf("expected 2 edges after merge, got %d", len(g1.Edges))
	}
}
