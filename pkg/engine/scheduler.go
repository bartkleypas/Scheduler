package engine

import (
	"context"
	"fmt"
	"time"
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
	config Config
	triggers chan IngestTrigger
	cancel   context.CancelFunc
}

// NewScheduler instantiates a new workspace indexing scheduler.
func NewScheduler(cfg Config) *Scheduler {
	if cfg.WatchDebounce == 0 {
		cfg.WatchDebounce = 500 * time.Millisecond
	}
	return &Scheduler{
		config:   cfg,
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

func (s *Scheduler) runWatcher(ctx context.Context) {
	// Stub: Will utilize fsnotify in Phase 2 implementation.
	fmt.Printf("[Watcher] Watching workspace: %s\n", s.config.WorkspacePath)
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
	fmt.Printf("[Ingestor] Ingest triggered by %s at %s\n", t.Type, t.Timestamp.Format(time.Kitchen))
	// Stub: Phase 2 will execute the AST & Debate parser and persist to SQLite here.
}
