package engine

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"scheduler/internal/ignore"
)

// runWatcher initializes fsnotify, registers directories, and streams debounced events.
func (s *Scheduler) runWatcher(ctx context.Context) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("[Watcher Error] Failed to create fsnotify watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	matcher := ignore.NewMatcher(s.config.WorkspacePath)
	absRoot, err := filepath.Abs(s.config.WorkspacePath)
	if err != nil {
		fmt.Printf("[Watcher Error] Failed to get absolute path: %v\n", err)
		return
	}

	// 1. Recursively add all non-ignored directories to watch list
	err = filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}

		if d.IsDir() {
			if matcher.IsIgnored(relPath) {
				return filepath.SkipDir
			}
			err = watcher.Add(path)
			if err != nil {
				fmt.Printf("[Watcher Error] Failed to watch directory %s: %v\n", path, err)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("[Watcher Error] Error walking directory tree: %v\n", err)
		return
	}

	fmt.Printf("[Watcher] Watching workspace: %s\n", absRoot)

	// State for debouncing file events
	var (
		debounceTimer *time.Timer
		debounceMu    sync.Mutex
		pendingPath   string
	)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Focus only on operations that modify files
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
				continue
			}

			relPath, err := filepath.Rel(absRoot, event.Name)
			if err != nil {
				continue
			}

			// Ignore events on ignored paths
			if matcher.IsIgnored(relPath) {
				continue
			}

			// Handle newly created directories dynamically
			if event.Op&fsnotify.Create != 0 {
				if isDir(event.Name) {
					err = watcher.Add(event.Name)
					if err == nil {
						fmt.Printf("[Watcher] Dynamically watching new folder: %s\n", relPath)
					}
					continue
				}
			}

			// Apply debouncing logic to group multiple rapid write events (e.g. editor temp saves)
			debounceMu.Lock()
			pendingPath = event.Name
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			debounceTimer = time.AfterFunc(s.config.WatchDebounce, func() {
				debounceMu.Lock()
				pathToSend := pendingPath
				debounceMu.Unlock()

				s.triggers <- IngestTrigger{
					Type:      TriggerFileChanged,
					Path:      pathToSend,
					Timestamp: time.Now(),
				}
			})
			debounceMu.Unlock()

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("[Watcher Error] fsnotify error: %v\n", err)

		case <-ctx.Done():
			fmt.Println("[Watcher] Watcher goroutine exiting...")
			return
		}
	}
}

// Helper function to check if a path is a directory
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
