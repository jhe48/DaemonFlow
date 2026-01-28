package task

import (
	"context"
	"log"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/jackyhe0402/daemonflow/internal/activity"
)

// TaskTracker monitors a task file for checkbox completions
type TaskTracker struct {
	filePath      string
	pollInterval  time.Duration
	previousTasks []Task

	// Listeners to notify on task completions
	listeners []activity.ActivityListener
	mu        sync.RWMutex

	// Callback for file content changes (for database sync)
	onFileChanged func(dir string)

	// Lifecycle control
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewTracker creates a new TaskTracker for the given file path
func NewTracker(filePath string, pollInterval time.Duration) (*TaskTracker, error) {
	// Resolve absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	// Default poll interval if not specified
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	return &TaskTracker{
		filePath:      absPath,
		pollInterval:  pollInterval,
		previousTasks: []Task{}, // Will be populated on first poll
		listeners:     make([]activity.ActivityListener, 0),
	}, nil
}

// AddListener adds a listener that will be notified of task completions
func (t *TaskTracker) AddListener(listener activity.ActivityListener) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.listeners = append(t.listeners, listener)
}

// SetOnFileChanged sets the callback to invoke when TASKS.md content changes.
// The callback receives the directory path containing the changed file.
func (t *TaskTracker) SetOnFileChanged(callback func(dir string)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onFileChanged = callback
}

// Start begins monitoring the task file for completions
func (t *TaskTracker) Start(ctx context.Context) error {
	// Create cancellable context
	ctx, t.cancel = context.WithCancel(ctx)

	// Do initial parse to populate previousTasks (ignore error if file doesn't exist yet)
	tasks, err := ParseFile(t.filePath)
	if err == nil {
		t.previousTasks = tasks
	}
	// If file doesn't exist, previousTasks remains empty - that's fine

	// Start poll loop goroutine
	t.wg.Add(1)
	go t.pollLoop(ctx)

	log.Printf("Task tracker started: monitoring %s (poll every %v)", t.filePath, t.pollInterval)
	return nil
}

// Stop stops the tracker
func (t *TaskTracker) Stop() {
	if t.cancel != nil {
		t.cancel()
		t.wg.Wait()
	}
	log.Println("Task tracker stopped")
}

// pollLoop periodically checks for task completions
func (t *TaskTracker) pollLoop(ctx context.Context) {
	defer t.wg.Done()

	ticker := time.NewTicker(t.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			t.checkForCompletions()
		}
	}
}

// checkForCompletions parses the current file state and detects newly completed tasks
func (t *TaskTracker) checkForCompletions() {
	// Parse current file state
	currentTasks, err := ParseFile(t.filePath)
	if err != nil {
		// File doesn't exist or can't be read - check if previous state had tasks
		if len(t.previousTasks) > 0 {
			// File was removed or became unreadable - this is a change
			t.notifyFileChanged()
		}
		t.previousTasks = []Task{}
		return
	}

	// Check if file content has changed
	hasChanges := t.hasChanges(currentTasks)

	// Build map of previous tasks by (line, text) for comparison
	// Key: "line:text", Value: Task
	prevMap := make(map[string]Task)
	for _, task := range t.previousTasks {
		key := strconv.Itoa(task.Line) + ":" + task.Text
		prevMap[key] = task
	}

	// Check each current task for new completions
	for _, current := range currentTasks {
		if !current.Completed {
			continue // Only interested in completed tasks
		}

		key := strconv.Itoa(current.Line) + ":" + current.Text
		prev, existed := prevMap[key]

		// Task is a new completion if:
		// 1. It didn't exist before (new task added as complete) - skip this case per spec
		// 2. It existed but was not completed - emit TaskComplete
		if existed && !prev.Completed {
			t.emitTaskComplete(current)
		}
	}

	// Update previous state for next comparison
	t.previousTasks = currentTasks

	// Notify callback if content changed (for database sync)
	if hasChanges {
		t.notifyFileChanged()
	}
}

// hasChanges compares current tasks to previous tasks to detect any changes.
// Returns true if tasks were added, removed, or modified.
func (t *TaskTracker) hasChanges(currentTasks []Task) bool {
	// Different count = definitely changed
	if len(currentTasks) != len(t.previousTasks) {
		return true
	}

	// Build map of previous tasks for comparison
	prevMap := make(map[string]Task)
	for _, task := range t.previousTasks {
		key := strconv.Itoa(task.Line) + ":" + task.Text
		prevMap[key] = task
	}

	// Check each current task
	for _, current := range currentTasks {
		key := strconv.Itoa(current.Line) + ":" + current.Text
		prev, existed := prevMap[key]
		if !existed {
			// New task
			return true
		}
		if prev.Completed != current.Completed {
			// Completion status changed
			return true
		}
	}

	return false
}

// notifyFileChanged calls the onFileChanged callback if set.
func (t *TaskTracker) notifyFileChanged() {
	t.mu.RLock()
	callback := t.onFileChanged
	t.mu.RUnlock()

	if callback != nil {
		callback(filepath.Dir(t.filePath))
	}
}

// emitTaskComplete creates and emits a TaskComplete activity
func (t *TaskTracker) emitTaskComplete(task Task) {
	act := activity.NewActivity(activity.TaskComplete, map[string]string{
		"task": task.Text,
		"line": strconv.Itoa(task.Line),
	})

	// Notify all listeners
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, listener := range t.listeners {
		listener.OnActivity(act)
	}

	log.Printf("Task completed: %s", task.Text)
}

// FilePath returns the path to the task file being monitored
func (t *TaskTracker) FilePath() string {
	return t.filePath
}
