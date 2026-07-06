package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
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
}

// Scheduler orchestrates the ingestion loop, coalescing interrupts and ticks.
type Scheduler struct {
	config       Config
	service      GraphService
	triggers     chan IngestTrigger
	cancel       context.CancelFunc
	lastModTimes map[string]time.Time
	lastModMu    sync.RWMutex
}

// NewScheduler instantiates a new workspace indexing scheduler.
func NewScheduler(cfg Config, service GraphService) *Scheduler {
	if cfg.WatchDebounce == 0 {
		cfg.WatchDebounce = 500 * time.Millisecond
	}
	return &Scheduler{
		config:       cfg,
		service:      service,
		triggers:     make(chan IngestTrigger, 100),
		lastModTimes: make(map[string]time.Time),
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

	if t.Type == TriggerFileChanged {
		absPath := t.Path
		relPath, err := filepath.Rel(s.config.WorkspacePath, absPath)
		if err != nil {
			relPath = filepath.Base(absPath)
		}
		relPath = filepath.ToSlash(relPath)

		info, err := os.Stat(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				// File was deleted or renamed, clean up its subgraph
				fmt.Printf("[Ingestor] File deleted, cleaning database graph: %s\n", relPath)
				if err := s.service.DeleteFileGraph(context.Background(), relPath); err != nil {
					fmt.Printf("[Ingestor Error] Failed to delete file graph: %v\n", err)
				}
				s.lastModMu.Lock()
				delete(s.lastModTimes, relPath)
				s.lastModMu.Unlock()
			}
			return
		}

		s.lastModMu.RLock()
		lastMod, ok := s.lastModTimes[relPath]
		s.lastModMu.RUnlock()

		if ok && !info.ModTime().After(lastMod) {
			// Unchanged, skip!
			return
		}

		fmt.Printf("[Ingestor] Incremental scan of file: %s\n", relPath)
		ext := filepath.Ext(absPath)

		var singleGraph *parser.Graph

		switch ext {
		case ".go":
			singleGraph, err = parser.ParseGoSource(absPath, s.config.WorkspacePath)
		case ".md":
			singleGraph, err = parser.ParseMarkdownDebate(absPath, s.config.WorkspacePath)
		default:
			fmt.Printf("[Ingestor] Skip: unsupported file extension for %s\n", relPath)
			return
		}

		if err != nil {
			fmt.Printf("[Ingestor Error] Failed to parse file: %v\n", err)
			return
		}

		// Transactional clean-and-save to ensure old symbols are pruned
		if err := s.service.DeleteFileGraph(context.Background(), relPath); err != nil {
			fmt.Printf("[Ingestor Warning] Failed to clean old file graph: %v\n", err)
		}

		fmt.Printf("Parsed modified file. Saving to graph storage...\n")
		if err := s.service.SaveGraph(context.Background(), singleGraph); err != nil {
			fmt.Printf("[Ingestor Error] Failed to save graph: %v\n", err)
			return
		}

		s.lastModMu.Lock()
		s.lastModTimes[relPath] = info.ModTime()
		s.lastModMu.Unlock()
		fmt.Println("[Ingestor] Successfully saved incremental update.")
		return
	}

	// Full workspace scan
	fmt.Println("[Ingestor] Starting full workspace scan...")
	workspaceFiles, err := WalkWorkspace(s.config.WorkspacePath)
	if err != nil {
		fmt.Printf("[Ingestor Error] Failed to walk workspace: %v\n", err)
		return
	}

	// Map of files currently present in workspace
	activeFiles := make(map[string]bool)
	for _, f := range workspaceFiles.GoFiles {
		rel, _ := filepath.Rel(s.config.WorkspacePath, f)
		activeFiles[filepath.ToSlash(rel)] = true
	}
	for _, f := range workspaceFiles.MarkdownFiles {
		rel, _ := filepath.Rel(s.config.WorkspacePath, f)
		activeFiles[filepath.ToSlash(rel)] = true
	}

	// 1. Prune Deleted Files from Database
	currentGraph, err := s.service.GetGraph(context.Background())
	if err == nil && currentGraph != nil {
		for _, node := range currentGraph.Nodes {
			if node.Type == "File" {
				if !activeFiles[node.ID] {
					fmt.Printf("[Ingestor] Pruning deleted file from database: %s\n", node.ID)
					if err := s.service.DeleteFileGraph(context.Background(), node.ID); err != nil {
						fmt.Printf("[Ingestor Error] Failed to delete file graph: %v\n", err)
					}
					s.lastModMu.Lock()
					delete(s.lastModTimes, node.ID)
					s.lastModMu.Unlock()
				}
			}
		}
	}

	// 2. Parse and Save Go Files if changed
	for _, goFile := range workspaceFiles.GoFiles {
		rel, _ := filepath.Rel(s.config.WorkspacePath, goFile)
		relPath := filepath.ToSlash(rel)

		info, err := os.Stat(goFile)
		if err != nil {
			continue
		}

		s.lastModMu.RLock()
		lastMod, ok := s.lastModTimes[relPath]
		s.lastModMu.RUnlock()

		if ok && !info.ModTime().After(lastMod) {
			continue
		}

		fmt.Printf("[Ingestor] Scanning modified file: %s\n", relPath)
		g, err := parser.ParseGoSource(goFile, s.config.WorkspacePath)
		if err != nil {
			fmt.Printf("[Ingestor Warning] Failed to parse Go file %s: %v\n", goFile, err)
			continue
		}

		// Transactional clean-and-save
		if err := s.service.DeleteFileGraph(context.Background(), relPath); err != nil {
			fmt.Printf("[Ingestor Warning] Failed to clean old file graph for %s: %v\n", relPath, err)
		}
		if err := s.service.SaveGraph(context.Background(), g); err != nil {
			fmt.Printf("[Ingestor Warning] Failed to save graph for %s: %v\n", relPath, err)
			continue
		}

		s.lastModMu.Lock()
		s.lastModTimes[relPath] = info.ModTime()
		s.lastModMu.Unlock()
	}

	// 3. Parse and Save Markdown Files if changed
	for _, mdFile := range workspaceFiles.MarkdownFiles {
		rel, _ := filepath.Rel(s.config.WorkspacePath, mdFile)
		relPath := filepath.ToSlash(rel)

		info, err := os.Stat(mdFile)
		if err != nil {
			continue
		}

		s.lastModMu.RLock()
		lastMod, ok := s.lastModTimes[relPath]
		s.lastModMu.RUnlock()

		if ok && !info.ModTime().After(lastMod) {
			continue
		}

		fmt.Printf("[Ingestor] Scanning modified file: %s\n", relPath)
		g, err := parser.ParseMarkdownDebate(mdFile, s.config.WorkspacePath)
		if err != nil {
			fmt.Printf("[Ingestor Warning] Failed to parse Markdown file %s: %v\n", mdFile, err)
			continue
		}

		// Transactional clean-and-save
		if err := s.service.DeleteFileGraph(context.Background(), relPath); err != nil {
			fmt.Printf("[Ingestor Warning] Failed to clean old file graph for %s: %v\n", relPath, err)
		}
		if err := s.service.SaveGraph(context.Background(), g); err != nil {
			fmt.Printf("[Ingestor Warning] Failed to save graph for %s: %v\n", relPath, err)
			continue
		}

		s.lastModMu.Lock()
		s.lastModTimes[relPath] = info.ModTime()
		s.lastModMu.Unlock()
	}

	// Get updated complete graph from SQLite for snapshot reporting
	updatedGraph, err := s.service.GetGraph(context.Background())
	if err != nil {
		fmt.Printf("[Ingestor Error] Failed to retrieve updated graph: %v\n", err)
		return
	}

	if updatedGraph != nil {
		fmt.Printf("[Ingestor] Ingest complete! Database contains %d total nodes and %d total edges.\n", len(updatedGraph.Nodes), len(updatedGraph.Edges))

		fmt.Println("--- EXTRACTED KNOWLEDGE GRAPH SNAPSHOT ---")
		for id, node := range updatedGraph.Nodes {
			fmt.Printf("  • Node [%s]: %s\n", node.Type, id)
		}
		for _, edge := range updatedGraph.Edges {
			fmt.Printf("  → Edge: (%s) -[%s]-> (%s)\n", edge.From, edge.Type, edge.To)
		}
		fmt.Println("------------------------------------------")
	} else {
		fmt.Println("[Ingestor] Ingest complete! Database graph is empty or nil.")
	}
}
