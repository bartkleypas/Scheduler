package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"scheduler/pkg/engine"
)

func main() {
	fmt.Println("====================================================")
	fmt.Println("🚀 Bootstrapping Workspace Knowledge Graph Scheduler")
	fmt.Println("====================================================")

	// Configure the scheduler with our resolved design decisions
	cfg := engine.Config{
		WorkspacePath: ".",
		DBPath:        ".knowledge/graph.db",
		ScanInterval:  5 * time.Second, // Scan every 5 seconds for demonstration
		EnableWatcher: true,            // Enable reactive FS watcher
	}

	scheduler := engine.NewScheduler(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS shutdown interrupts gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down gracefully...")
		cancel()
	}()

	// Start the unified scheduling loop
	if err := scheduler.Start(ctx); err != nil {
		fmt.Printf("❌ Failed to start scheduler: %v\n", err)
		os.Exit(1)
	}

	// Trigger an initial manual sync to index the workspace
	scheduler.SyncNow()

	// Run until context is cancelled by shutdown signal
	<-ctx.Done()
	fmt.Println("👋 Scheduler exited successfully.")
}
