package git

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/jackyhe0402/daemonflow/internal/activity"
)

// DefaultPollInterval is the default interval between git state checks
const DefaultPollInterval = 5 * time.Second

// Monitor watches a Git repository for changes
type Monitor struct {
	repo         *Repo
	pollInterval time.Duration

	// Last known state
	lastCommit string
	lastBranch string

	// Listeners to notify on changes
	listeners []activity.ActivityListener
	mu        sync.RWMutex

	// Lifecycle control
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewMonitor creates a new git monitor for the given repository
func NewMonitor(repo *Repo, pollInterval time.Duration) *Monitor {
	if pollInterval <= 0 {
		pollInterval = DefaultPollInterval
	}

	return &Monitor{
		repo:         repo,
		pollInterval: pollInterval,
		listeners:    make([]activity.ActivityListener, 0),
	}
}

// AddListener adds a listener that will be notified of git activities
func (m *Monitor) AddListener(listener activity.ActivityListener) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, listener)
}

// Start begins monitoring the repository for changes
func (m *Monitor) Start(ctx context.Context) error {
	// Initialize last known state
	if err := m.initializeState(); err != nil {
		return err
	}

	// Create cancellable context
	ctx, m.cancel = context.WithCancel(ctx)

	// Start polling goroutine
	m.wg.Add(1)
	go m.pollLoop(ctx)

	log.Printf("Git monitor started: polling every %v", m.pollInterval)
	return nil
}

// Stop stops the monitor
func (m *Monitor) Stop() {
	if m.cancel != nil {
		m.cancel()
		m.wg.Wait()
		log.Println("Git monitor stopped")
	}
}

// initializeState captures the current git state as baseline
func (m *Monitor) initializeState() error {
	commit, err := m.repo.GetHEAD()
	if err != nil {
		// No commits yet is OK
		log.Printf("Git monitor: could not get HEAD (repo may have no commits): %v", err)
		commit = ""
	}
	m.lastCommit = commit

	branch, err := m.repo.GetBranch()
	if err != nil {
		log.Printf("Git monitor: could not get branch: %v", err)
		branch = ""
	}
	m.lastBranch = branch

	log.Printf("Git monitor initialized: commit=%s, branch=%s", shortenHash(m.lastCommit), m.lastBranch)
	return nil
}

// pollLoop continuously checks for git changes
func (m *Monitor) pollLoop(ctx context.Context) {
	defer m.wg.Done()

	ticker := time.NewTicker(m.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkForChanges()
		}
	}
}

// checkForChanges detects and reports git state changes
func (m *Monitor) checkForChanges() {
	// Check for commit changes
	currentCommit, err := m.repo.GetHEAD()
	if err != nil {
		// Silent - repo might have no commits
		currentCommit = ""
	}

	if currentCommit != "" && currentCommit != m.lastCommit && m.lastCommit != "" {
		// New commit detected
		message, _ := m.repo.GetLastCommitMessage()

		act := activity.NewActivity(activity.GitCommit, map[string]string{
			"hash":    currentCommit,
			"message": message,
		})
		m.notifyListeners(act)
		log.Printf("Git activity: new commit %s - %s", shortenHash(currentCommit), message)
	}
	m.lastCommit = currentCommit

	// Check for branch changes
	currentBranch, err := m.repo.GetBranch()
	if err != nil {
		currentBranch = ""
	}

	if currentBranch != m.lastBranch && m.lastBranch != "" {
		act := activity.NewActivity(activity.GitBranchSwitch, map[string]string{
			"old_branch": m.lastBranch,
			"new_branch": currentBranch,
		})
		m.notifyListeners(act)
		log.Printf("Git activity: branch switch %s -> %s", m.lastBranch, currentBranch)
	}
	m.lastBranch = currentBranch
}

// notifyListeners sends an activity to all registered listeners
func (m *Monitor) notifyListeners(act activity.Activity) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, listener := range m.listeners {
		listener.OnActivity(act)
	}
}

// shortenHash returns first 7 chars of a git hash
func shortenHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}
