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
		// File doesn't exist or can't be read - clear state
		t.previousTasks = []Task{}
		return
	}

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
