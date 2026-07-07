package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"scheduler/pkg/parser"
)

// TriggerType defines what initiated the ingestion run.
type TriggerType string

const (
	TriggerFileChanged TriggerType = "FILE_CHANGED"
	TriggerPeriodicScan TriggerType = "PERIODIC_SCAN"
	TriggerManualSync   TriggerType = "MANUAL_SYNC"
)

// IngestTrigger carries metadata about the ingestion impulse.
type IngestTrigger struct {
	Type      TriggerType
	Path      string    // Specific file path for incremental changes (if applicable)
	Timestamp time.Time
}

// Config defines the scheduling and ingestion behavior.
type Config struct {
	WorkspacePath string
	DBPath        string
	ScanInterval  time.Duration // Duration for wall-clock interval ticks (0 to disable)
	EnableWatcher bool          // Enable reactive fsnotify file system watcher
	WatchDebounce time.Duration // Debounce time for rapid file modifications
	Verbose       bool          // Enable verbose console logging (e.g. graph snapshots)
}

// Scheduler orchestrates the ingestion loop, coalescing interrupts and ticks.
type Scheduler struct {
	config   Config
	service  GraphService
	triggers chan IngestTrigger
	cancel   context.CancelFunc
}

// NewScheduler instantiates a new workspace indexing scheduler.
func NewScheduler(cfg Config, service GraphService) *Scheduler {
	if cfg.WatchDebounce == 0 {
		cfg.WatchDebounce = 500 * time.Millisecond
	}
	if cfg.WorkspacePath != "" {
		if abs, err := filepath.Abs(cfg.WorkspacePath); err == nil {
			cfg.WorkspacePath = abs
		}
	}
	return &Scheduler{
		config:   cfg,
		service:  service,
		triggers: make(chan IngestTrigger, 100),
	}
}

// Start launches the background scheduler loop.
func (s *Scheduler) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// 1. Spawn Wall Clock Ticker if interval is set
	if s.config.ScanInterval > 0 {
		go s.runWallClock(ctx)
	}

	// 2. Spawn File System Watcher if enabled
	if s.config.EnableWatcher {
		go s.runWatcher(ctx)
	}

	// 3. Launch Coalescing Event loop
	go s.runEventLoop(ctx)

	return nil
}

// Trigger manual sync immediately.
func (s *Scheduler) SyncNow() {
	s.triggers <- IngestTrigger{
		Type:      TriggerManualSync,
		Timestamp: time.Now(),
	}
}

func (s *Scheduler) runWallClock(ctx context.Context) {
	ticker := time.NewTicker(s.config.ScanInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.triggers <- IngestTrigger{
				Type:      TriggerPeriodicScan,
				Timestamp: time.Now(),
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) runEventLoop(ctx context.Context) {
	fmt.Println("[Scheduler] Unified loop started...")
	for {
		select {
		case trigger := <-s.triggers:
			s.ingest(trigger)
		case <-ctx.Done():
			fmt.Println("[Scheduler] Unified loop stopped.")
			return
		}
	}
}

func (s *Scheduler) ingest(t IngestTrigger) {
	fmt.Printf("\n[Ingestor] Ingest triggered by %s at %s\n", t.Type, t.Timestamp.Format("15:04:05"))
	ctx := context.Background()

	if t.Type == TriggerFileChanged {
		absPath := t.Path
		relPath, err := filepath.Rel(s.config.WorkspacePath, absPath)
		if err != nil {
			relPath = filepath.Base(absPath)
		}
		relPath = filepath.ToSlash(relPath)

		_, err = os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				// File was deleted or renamed, clean up its subgraph and hash
				fmt.Printf("[Ingestor] File deleted, cleaning database graph and hash: %s\n", relPath)
				if err := s.service.DeleteFileGraph(ctx, relPath); err != nil {
					fmt.Printf("[Ingestor Error] Failed to delete file graph: %v\n", err)
				}
				if err := s.service.DeleteFileHash(ctx, relPath); err != nil {
					fmt.Printf("[Ingestor Error] Failed to delete file hash: %v\n", err)
				}
			}
			return
		}

		ext := filepath.Ext(absPath)
		var p parser.Parser
		for _, dp := range parser.DefaultParsers {
			if dp.CanHandle(ext) {
				p = dp
				break
			}
		}
		if p == nil {
			if s.config.Verbose {
				fmt.Printf("[Ingestor] Skip: unsupported file extension for %s\n", relPath)
			}
			return
		}

		if err := s.processFile(ctx, absPath, relPath, p); err != nil {
			fmt.Printf("[Ingestor Error] Failed to process file %s: %v\n", relPath, err)
		}
		return
	}

	// Full workspace scan
	if s.config.Verbose {
		fmt.Println("[Ingestor] Starting full workspace scan...")
	}
	workspaceFiles, err := WalkWorkspace(s.config.WorkspacePath)
	if err != nil {
		fmt.Printf("[Ingestor Error] Failed to walk workspace: %v\n", err)
		return
	}

	// Map of files currently present in workspace
	activeFiles := make(map[string]bool)
	allFiles := append(workspaceFiles.GoFiles, workspaceFiles.MarkdownFiles...)

	for _, f := range allFiles {
		rel, _ := filepath.Rel(s.config.WorkspacePath, f)
		activeFiles[filepath.ToSlash(rel)] = true
	}

	// 1. Prune Deleted Files from Database
	currentGraph, err := s.service.GetGraph(ctx)
	if err == nil && currentGraph != nil {
		for _, node := range currentGraph.Nodes {
			if node.Type == "File" {
				if !activeFiles[node.ID] {
					fmt.Printf("[Ingestor] Pruning deleted file from database: %s\n", node.ID)
					if err := s.service.DeleteFileGraph(ctx, node.ID); err != nil {
						fmt.Printf("[Ingestor Error] Failed to delete file graph: %v\n", err)
					}
					if err := s.service.DeleteFileHash(ctx, node.ID); err != nil {
						fmt.Printf("[Ingestor Error] Failed to delete file hash: %v\n", err)
					}
				}
			}
		}
	}

	// 2. Process all files on disk
	for _, f := range allFiles {
		rel, _ := filepath.Rel(s.config.WorkspacePath, f)
		relPath := filepath.ToSlash(rel)
		ext := filepath.Ext(f)

		var p parser.Parser
		for _, dp := range parser.DefaultParsers {
			if dp.CanHandle(ext) {
				p = dp
				break
			}
		}
		if p == nil {
			continue
		}

		if err := s.processFile(ctx, f, relPath, p); err != nil {
			fmt.Printf("[Ingestor Warning] Failed to process file %s: %v\n", relPath, err)
		}
	}

	// Get updated complete graph from SQLite for snapshot reporting
	updatedGraph, err := s.service.GetGraph(ctx)
	if err != nil {
		fmt.Printf("[Ingestor Error] Failed to retrieve updated graph: %v\n", err)
		return
	}

	if updatedGraph != nil {
		fmt.Printf("[Ingestor] Ingest complete! Database contains %d total nodes and %d total edges.\n", len(updatedGraph.Nodes), len(updatedGraph.Edges))

		if s.config.Verbose {
			fmt.Println("--- EXTRACTED KNOWLEDGE GRAPH SNAPSHOT ---")
			for id, node := range updatedGraph.Nodes {
				fmt.Printf("  • Node [%s]: %s\n", node.Type, id)
			}
			for _, edge := range updatedGraph.Edges {
				fmt.Printf("  → Edge: (%s) -[%s]-> (%s)\n", edge.From, edge.Type, edge.To)
			}
			fmt.Println("------------------------------------------")
		}
	} else {
		fmt.Println("[Ingestor] Ingest complete! Database graph is empty or nil.")
	}
}

func (s *Scheduler) processFile(ctx context.Context, absPath, relPath string, p parser.Parser) error {
	newHash, err := computeFileHash(absPath)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}

	oldHash, err := s.service.GetFileHash(ctx, relPath)
	if err == nil && oldHash == newHash {
		// Content unchanged, skip parsing and DB operations
		return nil
	}

	if s.config.Verbose {
		fmt.Printf("[Ingestor] Scanning modified file: %s\n", relPath)
	}

	g, err := p.Parse(absPath, s.config.WorkspacePath)
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	// Transactional clean-and-save
	if err := s.service.DeleteFileGraph(ctx, relPath); err != nil {
		fmt.Printf("[Ingestor Warning] Failed to clean old file graph for %s: %v\n", relPath, err)
	}
	if err := s.service.SaveGraph(ctx, g); err != nil {
		return fmt.Errorf("failed to save graph: %w", err)
	}

	if err := s.service.SetFileHash(ctx, relPath, newHash); err != nil {
		fmt.Printf("[Ingestor Warning] Failed to save hash for %s: %v\n", relPath, err)
	}

	return nil
}

func computeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
