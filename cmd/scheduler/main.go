package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"scheduler/internal/storage"
	"scheduler/pkg/api"
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

	// Initialize Storage
	store, err := storage.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		fmt.Printf("❌ Failed to initialize storage: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize the Graph Service
	service := storage.NewSQLiteGraphService(store)

	scheduler := engine.NewScheduler(cfg, service)

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

	// Initialize and launch the HTTP API Server
	apiServer := api.NewServer(scheduler, service, 8080)
	go func() {
		if err := apiServer.Start(); err != nil {
			fmt.Printf("❌ API Server failed to start: %v\n", err)
		}
	}()

	// Trigger an initial manual sync to index the workspace
	scheduler.SyncNow()

	// Run until context is cancelled by shutdown signal
	<-ctx.Done()
	
	// Graceful shutdown of the API Server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("❌ API Server graceful shutdown failed: %v\n", err)
	}

	fmt.Println("👋 Scheduler exited successfully.")
}
