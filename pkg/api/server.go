package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"scheduler/pkg/engine"
)

// Server represents the REST/JSON HTTP server for interacting with the Knowledge Graph.
type Server struct {
	scheduler *engine.Scheduler
	service   engine.GraphService
	srv       *http.Server
}

// NewServer instantiates a new REST API server.
func NewServer(scheduler *engine.Scheduler, service engine.GraphService, port int) *Server {
	mux := http.NewServeMux()
	s := &Server{
		scheduler: scheduler,
		service:   service,
	}

	// Register REST API endpoints (utilizing Go 1.22+ method-based pattern matching)
	mux.HandleFunc("GET /api/graph", s.handleGetGraph)
	mux.HandleFunc("POST /api/ingest", s.handleTriggerIngest)
	mux.HandleFunc("DELETE /api/nodes", s.handleDeleteNode)

	// Support preflight OPTIONS requests for CORS
	mux.HandleFunc("OPTIONS /api/graph", s.handleOptions)
	mux.HandleFunc("OPTIONS /api/ingest", s.handleOptions)
	mux.HandleFunc("OPTIONS /api/nodes", s.handleOptions)

	s.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	return s
}

// Start runs the HTTP server listening on the configured port.
func (s *Server) Start() error {
	fmt.Printf("[API Server] Listening on http://localhost%s\n", s.srv.Addr)
	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("API Server error: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("[API Server] Stopping server...")
	return s.srv.Shutdown(ctx)
}

// enableCors sets the appropriate CORS headers for cross-origin client interaction.
func (s *Server) enableCors(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) handleOptions(w http.ResponseWriter, r *http.Request) {
	s.enableCors(w)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetGraph(w http.ResponseWriter, r *http.Request) {
	s.enableCors(w)
	w.Header().Set("Content-Type", "application/json")

	graph, err := s.service.GetGraph(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if err := json.NewEncoder(w).Encode(graph); err != nil {
		fmt.Printf("[API Server Error] Failed to encode graph: %v\n", err)
	}
}

func (s *Server) handleTriggerIngest(w http.ResponseWriter, r *http.Request) {
	s.enableCors(w)
	w.Header().Set("Content-Type", "application/json")

	s.scheduler.SyncNow()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "Ingestion triggered"})
}

func (s *Server) handleDeleteNode(w http.ResponseWriter, r *http.Request) {
	s.enableCors(w)
	w.Header().Set("Content-Type", "application/json")

	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "missing id query parameter"})
		return
	}

	err := s.service.DeleteNode(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": fmt.Sprintf("node %s successfully deleted", id)})
}
