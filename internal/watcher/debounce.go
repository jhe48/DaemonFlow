package watcher

import (
	"sync"
	"time"
)

// FileEvent represents a file system event
type FileEvent struct {
	Path      string
	Operation string
}

// PendingEvent tracks a file event waiting to be flushed
type PendingEvent struct {
	Path      string
	Operation string
	FirstSeen time.Time
	LastSeen  time.Time
}

// Debouncer aggregates rapid file changes into batched callbacks
type Debouncer struct {
	pendingEvents  map[string]*PendingEvent
	debounceWindow time.Duration
	mu             sync.Mutex
	timer          *time.Timer
	callback       func(events []FileEvent)
}

// NewDebouncer creates a new Debouncer with the specified window
func NewDebouncer(debounceWindow time.Duration, callback func(events []FileEvent)) *Debouncer {
	if debounceWindow <= 0 {
		debounceWindow = 500 * time.Millisecond
	}
	return &Debouncer{
		pendingEvents:  make(map[string]*PendingEvent),
		debounceWindow: debounceWindow,
		callback:       callback,
	}
}

// Add adds an event to the debouncer
func (d *Debouncer) Add(event FileEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	if existing, ok := d.pendingEvents[event.Path]; ok {
		// Update existing event: keep FirstSeen, update LastSeen and Operation
		existing.LastSeen = now
		existing.Operation = event.Operation
	} else {
		// Add new pending event
		d.pendingEvents[event.Path] = &PendingEvent{
			Path:      event.Path,
			Operation: event.Operation,
			FirstSeen: now,
			LastSeen:  now,
		}
	}

	// Reset/start timer
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.debounceWindow, d.flush)
}

// Flush forces immediate flush of all pending events
func (d *Debouncer) Flush() {
	d.mu.Lock()
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	d.mu.Unlock()

	d.flush()
}

// flush sends all pending events to the callback and clears the map
func (d *Debouncer) flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.pendingEvents) == 0 {
		return
	}

	// Collect all events
	events := make([]FileEvent, 0, len(d.pendingEvents))
	for _, pending := range d.pendingEvents {
		events = append(events, FileEvent{
			Path:      pending.Path,
			Operation: pending.Operation,
		})
	}

	// Clear pending events
	d.pendingEvents = make(map[string]*PendingEvent)

	// Call callback outside of lock
	if d.callback != nil {
		d.callback(events)
	}
}

// PendingCount returns the number of pending events (for testing)
func (d *Debouncer) PendingCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.pendingEvents)
}
