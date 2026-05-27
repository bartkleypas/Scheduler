package engine

import (
	"context"
	"fmt"
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

	if t.Type == TriggerFileChanged {
		// Incremental, surgical single-file parse
		fmt.Printf("[Ingestor] Incremental scan of file: %s\n", t.Path)
		ext := filepath.Ext(t.Path)

		var singleGraph *parser.Graph
		var err error

		switch ext {
		case ".go":
			singleGraph, err = parser.ParseGoSource(t.Path, s.config.WorkspacePath)
		case ".md":
			singleGraph, err = parser.ParseMarkdownDebate(t.Path, s.config.WorkspacePath)
		default:
			fmt.Printf("[Ingestor] Skip: unsupported file extension for %s\n", t.Path)
			return
		}

		if err != nil {
			fmt.Printf("[Ingestor Error] Failed to parse file: %v\n", err)
			return
		}

		fmt.Printf("Parsed 1 file. Saving to graph storage...\n")
		if err := s.service.SaveGraph(context.Background(), singleGraph); err != nil {
			fmt.Printf("[Ingestor Error] Failed to save graph: %v\n", err)
			return
		}
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

	fullGraph := parser.NewGraph()

	// Parse all Go files
	for _, goFile := range workspaceFiles.GoFiles {
		g, err := parser.ParseGoSource(goFile, s.config.WorkspacePath)
		if err != nil {
			fmt.Printf("[Ingestor Warning] Failed to parse Go file %s: %v\n", goFile, err)
			continue
		}
		fullGraph.Merge(g)
	}

	// Parse all Markdown files
	for _, mdFile := range workspaceFiles.MarkdownFiles {
		g, err := parser.ParseMarkdownDebate(mdFile, s.config.WorkspacePath)
		if err != nil {
			fmt.Printf("[Ingestor Warning] Failed to parse Markdown file %s: %v\n", mdFile, err)
			continue
		}
		fullGraph.Merge(g)
	}

	fmt.Printf("[Ingestor] Ingest complete! Discovered %d total nodes and %d total edges.\n", len(fullGraph.Nodes), len(fullGraph.Edges))
	
	fmt.Println("[Ingestor] Saving full workspace graph to storage...")
	if err := s.service.SaveGraph(context.Background(), fullGraph); err != nil {
		fmt.Printf("[Ingestor Error] Failed to save full workspace graph: %v\n", err)
		return
	}
	fmt.Println("[Ingestor] Successfully saved full workspace graph.")

	fmt.Println("--- EXTRACTED KNOWLEDGE GRAPH SNAPSHOT ---")
	for id, node := range fullGraph.Nodes {
		fmt.Printf("  • Node [%s]: %s\n", node.Type, id)
	}
	for _, edge := range fullGraph.Edges {
		fmt.Printf("  → Edge: (%s) -[%s]-> (%s)\n", edge.From, edge.Type, edge.To)
	}
	fmt.Println("------------------------------------------")
}
