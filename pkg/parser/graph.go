package parser

// Node represents a vertex in the Knowledge Graph.
type Node struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`       // "File", "Symbol", "Debate", "Option", "Decision"
	Properties map[string]interface{} `json:"properties"` // Dynamic key-value attributes
}

// Edge represents a directed relation between two vertices in the Knowledge Graph.
type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // "IMPORTS", "DEFINED_IN", "TARGETS", "PROPOSES", "APPROVED", "RESOLVES"
}

// Graph holds the extracted workspace elements.
type Graph struct {
	Nodes map[string]Node `json:"nodes"`
	Edges []Edge          `json:"edges"`
}

// NewGraph instantiates an empty knowledge graph storage container.
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]Node),
		Edges: make([]Edge, 0),
	}
}

// AddNode registers a node in the graph, overwriting if ID matches.
func (g *Graph) AddNode(node Node) {
	g.Nodes[node.ID] = node
}

// AddEdge registers a directed connection in the graph.
func (g *Graph) AddEdge(from, to, edgeType string) {
	g.Edges = append(g.Edges, Edge{From: from, To: to, Type: edgeType})
}

// Merge merges nodes and edges from another graph into this one.
func (g *Graph) Merge(other *Graph) {
	if other == nil {
		return
	}
	for id, node := range other.Nodes {
		g.Nodes[id] = node
	}
	g.Edges = append(g.Edges, other.Edges...)
}
