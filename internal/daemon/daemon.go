package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/jackyhe0402/daemonflow/internal/activity"
	"github.com/jackyhe0402/daemonflow/internal/clock"
	"github.com/jackyhe0402/daemonflow/internal/config"
	"github.com/jackyhe0402/daemonflow/internal/git"
	"github.com/jackyhe0402/daemonflow/internal/graveyard"
	"github.com/jackyhe0402/daemonflow/internal/ipc"
	"github.com/jackyhe0402/daemonflow/internal/store"
	"github.com/jackyhe0402/daemonflow/internal/task"
	"github.com/jackyhe0402/daemonflow/internal/watcher"
)

// MaxRecentActivities is the maximum number of activities to keep in memory
const MaxRecentActivities = 50

// Daemon manages the DaemonFlow background process
type Daemon struct {
	Running      bool
	PIDFile      string
	DataDir      string
	ConfigPath   string
	Foreground   bool
	Config       *config.Config
	ipcServer    *ipc.Server
	gitMonitor   *git.Monitor
	fileWatcher  *watcher.Watcher
	taskTracker  *task.TaskTracker
	clock        *clock.Clock
	graveyard    *graveyard.Graveyard
	store        *store.Store
	startTime    time.Time
	shutdownChan chan struct{}

	// Activity tracking
	recentActivities []activity.Activity
	activityMu       sync.RWMutex
}

// New creates a new Daemon instance with default paths
func New() *Daemon {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	dataDir := filepath.Join(homeDir, ".daemonflow")

	return &Daemon{
		DataDir: dataDir,
		PIDFile: filepath.Join(dataDir, "daemonflow.pid"),
	}
}

// Start starts the daemon
func (d *Daemon) Start() error {
	// Check if already running
	if d.isRunning() {
		return fmt.Errorf("daemon is already running")
	}

	// If not in foreground mode, fork to background
	if !d.Foreground {
		return d.startBackground()
	}

	// Foreground mode - run directly
	return d.runForeground()
}

// startBackground spawns the daemon as a background process
func (d *Daemon) startBackground() error {
	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Build command with --foreground flag
	args := []string{"start", "--foreground"}
	if d.ConfigPath != "" {
		args = append(args, "--config", d.ConfigPath)
	}

	cmd := exec.Command(executable, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Start in background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("DaemonFlow daemon started (PID: %d)\n", cmd.Process.Pid)
	return nil
}

// runForeground runs the daemon in foreground mode
func (d *Daemon) runForeground() error {
	// Load configuration
	cfg, err := config.LoadConfig(d.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	d.Config = cfg

	// Create data directory if not exists
	if err := os.MkdirAll(d.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open SQLite store
	dbPath := filepath.Join(d.DataDir, "daemonflow.db")
	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open SQLite store: %w", err)
	}
	d.store = s
	log.Printf("SQLite store opened: %s", dbPath)

	// Write PID file
	pid := os.Getpid()
	if err := os.WriteFile(d.PIDFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	defer os.Remove(d.PIDFile)

	// Record start time
	d.startTime = time.Now()
	d.shutdownChan = make(chan struct{})
	d.recentActivities = make([]activity.Activity, 0, MaxRecentActivities)

	// Create graveyard for death logging
	d.graveyard = graveyard.NewGraveyard(d.DataDir)
	// Load existing death records (ignore error if file doesn't exist)
	if err := d.graveyard.LoadRecords(); err != nil {
		log.Printf("Warning: failed to load graveyard records: %v", err)
	}

	// Create and start Freedom Clock
	d.clock = clock.NewClock(&d.Config.Earning)

	// Wire death callback to log deaths to graveyard
	d.clock.OnDeath = func(overtimeSeconds int, sessionEarned int) {
		record := graveyard.DeathRecord{
			Timestamp:       time.Now(),
			OvertimeSeconds: -overtimeSeconds, // Convert negative to positive
			SessionEarned:   sessionEarned,
			Cause:           "overtime",
		}
		if err := d.graveyard.LogDeath(record); err != nil {
			// Log error but don't crash - death logging is not critical
			log.Printf("Failed to log death: %v", err)
		} else {
			log.Printf("Pet died after %dm %ds of overtime (total deaths: %d)",
				record.OvertimeSeconds/60, record.OvertimeSeconds%60,
				d.graveyard.GetDeathCount())
		}
	}

	d.clock.Start()

	// Start IPC server
	d.ipcServer = ipc.NewServer(d.Config.SocketPath, d)
	if err := d.ipcServer.Start(); err != nil {
		return fmt.Errorf("failed to start IPC server: %w", err)
	}
	defer d.ipcServer.Stop()

	// Start git monitor if watch dir is a git repo
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := d.startGitMonitor(ctx); err != nil {
		log.Printf("Git monitor not started: %v", err)
	}

	// Start file watcher if enabled
	if err := d.startFileWatcher(ctx); err != nil {
		log.Printf("File watcher not started: %v", err)
	}

	// Start task tracker if enabled
	if err := d.startTaskTracker(ctx); err != nil {
		log.Printf("Task tracker not started: %v", err)
	}

	d.Running = true
	log.Printf("DaemonFlow daemon running (PID: %d)", pid)
	log.Printf("Config: watch_dir=%s, log_level=%s", d.Config.WatchDir, d.Config.LogLevel)
	log.Printf("IPC socket: %s", d.Config.SocketPath)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main loop
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down...", sig)
			d.Running = false
			d.stopClock()
			d.stopFileWatcher()
			d.stopTaskTracker()
			d.stopGitMonitor()
			d.closeStore()
			return nil
		case <-d.shutdownChan:
			log.Printf("Received shutdown request via IPC, shutting down...")
			d.Running = false
			d.stopClock()
			d.stopFileWatcher()
			d.stopTaskTracker()
			d.stopGitMonitor()
			d.closeStore()
			return nil
		case <-ticker.C:
			log.Println("daemon running")
		}
	}
}

// startGitMonitor initializes and starts the git monitor if applicable
func (d *Daemon) startGitMonitor(ctx context.Context) error {
	// Find git root from watch directory
	gitRoot, err := git.FindGitRoot(d.Config.WatchDir)
	if err != nil {
		return fmt.Errorf("watch directory is not a git repository: %w", err)
	}

	// Create repo and monitor
	repo := git.NewRepo(gitRoot)
	if !repo.IsGitRepo() {
		return fmt.Errorf("not a valid git repository: %s", gitRoot)
	}

	d.gitMonitor = git.NewMonitor(repo, d.Config.Git.PollInterval)
	d.gitMonitor.AddListener(d)

	if err := d.gitMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start git monitor: %w", err)
	}

	log.Printf("Git monitor watching: %s (poll every %v)", gitRoot, d.Config.Git.PollInterval)
	return nil
}

// stopGitMonitor stops the git monitor if running
func (d *Daemon) stopGitMonitor() {
	if d.gitMonitor != nil {
		d.gitMonitor.Stop()
	}
}

// startFileWatcher initializes and starts the file watcher if enabled
func (d *Daemon) startFileWatcher(ctx context.Context) error {
	if !d.Config.Watcher.Enabled {
		log.Printf("File watcher disabled in config")
		return nil
	}

	// Create ignore matcher with default + config patterns
	ignore := watcher.NewIgnoreMatcherWithPatterns(d.Config.Watcher.IgnorePatterns)

	// Create watcher with config values
	fw, err := watcher.NewWatcher(d.Config.WatchDir, ignore, d.Config.Watcher.DebounceWindow)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Add daemon as listener (OnActivity already implemented)
	fw.AddListener(d)

	// Start watcher
	if err := fw.Start(ctx); err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	d.fileWatcher = fw
	log.Printf("File watcher monitoring: %s (debounce %v)", d.Config.WatchDir, d.Config.Watcher.DebounceWindow)
	return nil
}

// stopFileWatcher stops the file watcher if running
func (d *Daemon) stopFileWatcher() {
	if d.fileWatcher != nil {
		d.fileWatcher.Stop()
	}
}

// startTaskTracker initializes and starts the task tracker if enabled
func (d *Daemon) startTaskTracker(ctx context.Context) error {
	if !d.Config.Task.Enabled {
		log.Printf("Task tracker disabled in config")
		return nil
	}

	// Resolve task file path: if relative, join with watch_dir
	taskFilePath := d.Config.Task.FilePath
	if !filepath.IsAbs(taskFilePath) {
		taskFilePath = filepath.Join(d.Config.WatchDir, taskFilePath)
	}

	// Create task tracker
	tt, err := task.NewTracker(taskFilePath, d.Config.Task.PollInterval)
	if err != nil {
		return fmt.Errorf("failed to create task tracker: %w", err)
	}

	// Add daemon as listener (d implements ActivityListener)
	tt.AddListener(d)

	// Start tracker
	if err := tt.Start(ctx); err != nil {
		return fmt.Errorf("failed to start task tracker: %w", err)
	}

	d.taskTracker = tt
	log.Printf("Task tracker monitoring: %s (poll every %v)", taskFilePath, d.Config.Task.PollInterval)
	return nil
}

// stopTaskTracker stops the task tracker if running
func (d *Daemon) stopTaskTracker() {
	if d.taskTracker != nil {
		d.taskTracker.Stop()
	}
}

// stopClock stops the Freedom Clock if running
func (d *Daemon) stopClock() {
	if d.clock != nil {
		d.clock.Stop()
	}
}

// closeStore closes the SQLite store if open
func (d *Daemon) closeStore() {
	if d.store != nil {
		if err := d.store.Close(); err != nil {
			log.Printf("Error closing SQLite store: %v", err)
		} else {
			log.Println("SQLite store closed")
		}
	}
}

// GetStore returns the SQLite store instance.
// May be nil if daemon is not running.
func (d *Daemon) GetStore() *store.Store {
	return d.store
}

// OnActivity implements activity.ActivityListener
func (d *Daemon) OnActivity(act activity.Activity) {
	d.activityMu.Lock()
	defer d.activityMu.Unlock()

	// Add to recent activities
	d.recentActivities = append(d.recentActivities, act)

	// Trim to max size
	if len(d.recentActivities) > MaxRecentActivities {
		d.recentActivities = d.recentActivities[len(d.recentActivities)-MaxRecentActivities:]
	}

	// Forward to Freedom Clock for earning calculation
	if d.clock != nil {
		d.clock.OnActivity(act)
	}
}

// GetRecentActivities returns the most recent N activities
func (d *Daemon) GetRecentActivities(limit int) []activity.Activity {
	d.activityMu.RLock()
	defer d.activityMu.RUnlock()

	if limit <= 0 || limit > len(d.recentActivities) {
		limit = len(d.recentActivities)
	}

	// Return most recent activities (last N items)
	start := len(d.recentActivities) - limit
	if start < 0 {
		start = 0
	}

	result := make([]activity.Activity, limit)
	copy(result, d.recentActivities[start:])
	return result
}

// GetRecentActivitiesData returns recent activities in IPC-friendly format
func (d *Daemon) GetRecentActivitiesData(limit int) []ipc.ActivityData {
	activities := d.GetRecentActivities(limit)
	data := make([]ipc.ActivityData, len(activities))
	for i, act := range activities {
		data[i] = ipc.ActivityData{
			Type:      string(act.Type),
			Timestamp: act.Timestamp.Format(time.RFC3339),
			Details:   act.Details,
		}
	}
	return data
}

// GetUptime returns the number of seconds the daemon has been running
func (d *Daemon) GetUptime() int64 {
	return int64(time.Since(d.startTime).Seconds())
}

// GetWatchDir returns the configured watch directory
func (d *Daemon) GetWatchDir() string {
	if d.Config != nil {
		return d.Config.WatchDir
	}
	return ""
}

// RequestShutdown requests a graceful daemon shutdown
func (d *Daemon) RequestShutdown() {
	if d.shutdownChan != nil {
		close(d.shutdownChan)
	}
}

// GetSocketPath returns the IPC socket path
func (d *Daemon) GetSocketPath() string {
	if d.Config != nil {
		return d.Config.SocketPath
	}
	// Return default path if config not loaded
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".daemonflow", "daemonflow.sock")
}

// GetClockState returns the current clock state ("working", "break", "overtime")
func (d *Daemon) GetClockState() string {
	if d.clock != nil {
		return string(d.clock.GetState())
	}
	return "working"
}

// GetEarnedSeconds returns the current earned seconds balance
func (d *Daemon) GetEarnedSeconds() int {
	if d.clock != nil {
		return d.clock.GetEarnedSeconds()
	}
	return 0
}

// GetSessionEarned returns total seconds earned this session
func (d *Daemon) GetSessionEarned() int {
	if d.clock != nil {
		return d.clock.GetSessionEarned()
	}
	return 0
}

// StartBreak transitions the clock to break state
func (d *Daemon) StartBreak() (previousState, newState string) {
	if d.clock != nil {
		return d.clock.StartBreak()
	}
	return "working", "working"
}

// EndBreak transitions the clock back to working state
func (d *Daemon) EndBreak() (previousState, newState string) {
	if d.clock != nil {
		return d.clock.EndBreak()
	}
	return "working", "working"
}

// GetStreakInfo returns streak statistics from the graveyard
func (d *Daemon) GetStreakInfo() (currentStreak, longestStreak, totalDeaths int) {
	if d.graveyard == nil {
		return 1, 1, 0 // Default: 1 day streak if no graveyard
	}
	info := d.graveyard.GetStreakInfo()
	return info.CurrentStreak, info.LongestStreak, d.graveyard.GetDeathCount()
}

// Resurrect attempts to revive the pet after death
// Returns an error if the pet is not dead
func (d *Daemon) Resurrect() (*ipc.ResurrectResponse, error) {
	// Check if pet is actually dead
	if d.graveyard == nil || !d.graveyard.IsDead() {
		return &ipc.ResurrectResponse{
			Success:  false,
			Message:  "Pet is not dead",
			NewState: d.GetClockState(),
		}, nil
	}

	// Log the resurrection
	if err := d.graveyard.LogResurrection(); err != nil {
		return nil, fmt.Errorf("failed to log resurrection: %w", err)
	}

	// Reset the clock (forfeit all earned time as recovery cost)
	if d.clock != nil {
		d.clock.Reset()
	}

	log.Printf("Pet resurrected! All earned time reset (total resurrections: %d)",
		d.graveyard.GetResurrectionCount())

	return &ipc.ResurrectResponse{
		Success:  true,
		Message:  "Pet resurrected! All earned time reset.",
		NewState: d.GetClockState(),
	}, nil
}

// Stop stops the running daemon
func (d *Daemon) Stop() error {
	pid, err := d.readPID()
	if err != nil {
		return fmt.Errorf("daemon is not running (no PID file)")
	}

	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Send SIGTERM
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Check if process doesn't exist
		if err == os.ErrProcessDone {
			os.Remove(d.PIDFile)
			return fmt.Errorf("daemon was not running (stale PID file removed)")
		}
		return fmt.Errorf("failed to send signal: %w", err)
	}

	fmt.Printf("Sent stop signal to DaemonFlow daemon (PID: %d)\n", pid)

	// Wait briefly for process to terminate
	time.Sleep(500 * time.Millisecond)

	// Check if still running
	if d.processExists(pid) {
		fmt.Println("Daemon is still shutting down...")
	} else {
		fmt.Println("Daemon stopped successfully")
	}

	return nil
}

// Status returns the current daemon status
func (d *Daemon) Status() (bool, int, error) {
	pid, err := d.readPID()
	if err != nil {
		return false, 0, nil // Not running, no error
	}

	if d.processExists(pid) {
		return true, pid, nil // Running
	}

	// Stale PID file
	os.Remove(d.PIDFile)
	return false, 0, nil
}

// isRunning checks if daemon is already running
func (d *Daemon) isRunning() bool {
	running, _, _ := d.Status()
	return running
}

// readPID reads the PID from the PID file
func (d *Daemon) readPID() (int, error) {
	data, err := os.ReadFile(d.PIDFile)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in file: %w", err)
	}
	return pid, nil
}

// processExists checks if a process with given PID exists
func (d *Daemon) processExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds, so we need to send signal 0 to check
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
