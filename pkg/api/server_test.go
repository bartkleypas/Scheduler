package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"scheduler/pkg/engine"
	"scheduler/pkg/parser"
)

type mockGraphService struct {
	graph       *parser.Graph
	deleteID    string
	deleteCount int
	saveCount   int
}

func (m *mockGraphService) SaveGraph(ctx context.Context, g *parser.Graph) error {
	m.saveCount++
	m.graph = g
	return nil
}

func (m *mockGraphService) GetGraph(ctx context.Context) (*parser.Graph, error) {
	return m.graph, nil
}

func (m *mockGraphService) DeleteNode(ctx context.Context, id string) error {
	m.deleteID = id
	m.deleteCount++
	return nil
}

func (m *mockGraphService) Close() error {
	return nil
}

func TestGetGraph(t *testing.T) {
	g := parser.NewGraph()
	g.AddNode(parser.Node{
		ID:         "test-node",
		Type:       "File",
		Properties: map[string]interface{}{"filepath": "test.go"},
	})

	mockSvc := &mockGraphService{graph: g}
	cfg := engine.Config{
		WorkspacePath: ".",
		ScanInterval:  0,
		EnableWatcher: false,
	}
	scheduler := engine.NewScheduler(cfg, mockSvc)
	server := NewServer(scheduler, mockSvc, 8080)

	req := httptest.NewRequest("GET", "/api/graph", nil)
	rr := httptest.NewRecorder()

	server.srv.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify CORS header
	if origin := rr.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("CORS header Access-Control-Allow-Origin not set correctly: got %v want *", origin)
	}

	var responseGraph parser.Graph
	if err := json.Unmarshal(rr.Body.Bytes(), &responseGraph); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if _, exists := responseGraph.Nodes["test-node"]; !exists {
		t.Errorf("expected node 'test-node' not found in response")
	}
}

func TestTriggerIngest(t *testing.T) {
	mockSvc := &mockGraphService{}
	cfg := engine.Config{
		WorkspacePath: ".",
		ScanInterval:  0,
		EnableWatcher: false,
	}
	scheduler := engine.NewScheduler(cfg, mockSvc)
	server := NewServer(scheduler, mockSvc, 8080)

	req := httptest.NewRequest("POST", "/api/ingest", nil)
	rr := httptest.NewRecorder()

	server.srv.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusAccepted)
	}

	expected := `{"status":"Ingestion triggered"}`
	body := strings.TrimSpace(rr.Body.String())
	if body != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", body, expected)
	}
}

func TestDeleteNode(t *testing.T) {
	mockSvc := &mockGraphService{}
	cfg := engine.Config{
		WorkspacePath: ".",
		ScanInterval:  0,
		EnableWatcher: false,
	}
	scheduler := engine.NewScheduler(cfg, mockSvc)
	server := NewServer(scheduler, mockSvc, 8080)

	// Test missing ID
	reqMissing := httptest.NewRequest("DELETE", "/api/nodes", nil)
	rrMissing := httptest.NewRecorder()
	server.srv.Handler.ServeHTTP(rrMissing, reqMissing)
	if status := rrMissing.Code; status != http.StatusBadRequest {
		t.Errorf("expected StatusBadRequest when missing ID; got %v", status)
	}

	// Test valid delete
	req := httptest.NewRequest("DELETE", "/api/nodes?id=node-to-delete", nil)
	rr := httptest.NewRecorder()

	server.srv.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if mockSvc.deleteCount != 1 {
		t.Errorf("expected delete count 1; got %d", mockSvc.deleteCount)
	}

	if mockSvc.deleteID != "node-to-delete" {
		t.Errorf("expected delete ID 'node-to-delete'; got %s", mockSvc.deleteID)
	}

	expected := `{"status":"node node-to-delete successfully deleted"}`
	body := strings.TrimSpace(rr.Body.String())
	if body != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", body, expected)
	}
}

func TestOptionsPreflight(t *testing.T) {
	mockSvc := &mockGraphService{}
	cfg := engine.Config{
		WorkspacePath: ".",
		ScanInterval:  0,
		EnableWatcher: false,
	}
	scheduler := engine.NewScheduler(cfg, mockSvc)
	server := NewServer(scheduler, mockSvc, 8080)

	req := httptest.NewRequest("OPTIONS", "/api/graph", nil)
	rr := httptest.NewRecorder()

	server.srv.Handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for OPTIONS preflight: got %v want %v", status, http.StatusOK)
	}

	if methods := rr.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "GET") {
		t.Errorf("expected CORS methods to allow GET; got %s", methods)
	}
}
