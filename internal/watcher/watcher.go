package watcher

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/jackyhe0402/daemonflow/internal/activity"
)

// Watcher monitors a directory for file changes
type Watcher struct {
	root    string
	watcher *fsnotify.Watcher
	ignore  *IgnoreMatcher

	// Listeners to notify on changes
	listeners []activity.ActivityListener
	mu        sync.RWMutex

	// Lifecycle control
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewWatcher creates a new file watcher for the given root directory
func NewWatcher(root string, ignore *IgnoreMatcher) (*Watcher, error) {
	// Resolve absolute path
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	// Verify directory exists
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, os.ErrInvalid
	}

	if ignore == nil {
		ignore = NewIgnoreMatcher()
	}

	return &Watcher{
		root:      absRoot,
		ignore:    ignore,
		listeners: make([]activity.ActivityListener, 0),
	}, nil
}

// AddListener adds a listener that will be notified of file changes
func (w *Watcher) AddListener(listener activity.ActivityListener) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.listeners = append(w.listeners, listener)
}

// Start begins watching the directory for changes
func (w *Watcher) Start(ctx context.Context) error {
	// Create fsnotify watcher
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	w.watcher = fsWatcher

	// Add all directories recursively
	if err := w.addDirectoriesRecursively(w.root); err != nil {
		w.watcher.Close()
		return err
	}

	// Create cancellable context
	ctx, w.cancel = context.WithCancel(ctx)

	// Start event handling goroutine
	w.wg.Add(1)
	go w.eventLoop(ctx)

	log.Printf("File watcher started: watching %s", w.root)
	return nil
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	if w.cancel != nil {
		w.cancel()
		w.wg.Wait()
	}
	if w.watcher != nil {
		w.watcher.Close()
	}
	log.Println("File watcher stopped")
}

// addDirectoriesRecursively adds all non-ignored directories to the watcher
func (w *Watcher) addDirectoriesRecursively(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log but continue on errors
			log.Printf("File watcher: error accessing %s: %v", path, err)
			return nil
		}

		// Get relative path for ignore matching
		relPath, err := filepath.Rel(w.root, path)
		if err != nil {
			relPath = path
		}

		// Skip if ignored
		if w.ignore.Match(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Watch directories
		if info.IsDir() {
			if err := w.watcher.Add(path); err != nil {
				log.Printf("File watcher: could not watch %s: %v", path, err)
			}
		}

		return nil
	})
}

// eventLoop handles fsnotify events
func (w *Watcher) eventLoop(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			w.handleEvent(event)

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("File watcher error: %v", err)
		}
	}
}

// handleEvent processes a single fsnotify event
func (w *Watcher) handleEvent(event fsnotify.Event) {
	// Get relative path for ignore matching and activity details
	relPath, err := filepath.Rel(w.root, event.Name)
	if err != nil {
		relPath = event.Name
	}

	// Skip ignored paths
	if w.ignore.Match(relPath) {
		return
	}

	// Determine operation type
	var operation string
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		operation = "create"
		// If a new directory is created, add it to the watcher
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			if err := w.watcher.Add(event.Name); err != nil {
				log.Printf("File watcher: could not watch new directory %s: %v", event.Name, err)
			}
		}
	case event.Op&fsnotify.Write == fsnotify.Write:
		operation = "modify"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		operation = "delete"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		operation = "delete" // Renamed files appear as delete + create
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		// Ignore permission changes
		return
	default:
		return
	}

	// Emit FileChange activity
	act := activity.NewActivity(activity.FileChange, map[string]string{
		"path":      relPath,
		"operation": operation,
	})
	w.notifyListeners(act)
	log.Printf("File activity: %s %s", operation, relPath)
}

// notifyListeners sends an activity to all registered listeners
func (w *Watcher) notifyListeners(act activity.Activity) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for _, listener := range w.listeners {
		listener.OnActivity(act)
	}
}

// Root returns the root directory being watched
func (w *Watcher) Root() string {
	return w.root
}
