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
	"github.com/jackyhe0402/daemonflow/internal/brain"
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
	Running       bool
	PIDFile       string
	DataDir       string
	ConfigPath    string
	Foreground    bool
	Config        *config.Config
	ipcServer     *ipc.Server
	gitMonitor    *git.Monitor
	fileWatcher   *watcher.Watcher
	taskTrackers  []*task.TaskTracker
	taskSync      *task.TaskSync
	clock         *clock.Clock
	graveyard     *graveyard.Graveyard
	store         *store.Store
	brainExecutor *brain.Executor
	startTime     time.Time
	shutdownChan  chan struct{}

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

	// Initialize brain executor (optional - Python may not be available)
	if err := d.initBrainExecutor(); err != nil {
		log.Printf("Warning: Python brain not available: %v", err)
		// Brain executor is optional - daemon can still function without it
		d.brainExecutor = nil
	}

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

	// Wire death callback to log deaths to graveyard and sync to SQLite
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
		// Sync graveyard stats to SQLite after death
		d.syncGraveyardToStore()
	}

	// Initial sync of graveyard stats to SQLite
	d.syncGraveyardToStore()

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

	// Start task trackers if enabled (one per watch directory)
	if err := d.startTaskTrackers(ctx); err != nil {
		log.Printf("Task trackers not started: %v", err)
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
			d.stopTaskTrackers()
			d.stopGitMonitor()
			d.closeStore()
			return nil
		case <-d.shutdownChan:
			log.Printf("Received shutdown request via IPC, shutting down...")
			d.Running = false
			d.stopClock()
			d.stopFileWatcher()
			d.stopTaskTrackers()
			d.stopGitMonitor()
			d.closeStore()
			return nil
		case <-ticker.C:
			log.Println("daemon running")
		}
	}
}

// startGitMonitor initializes and starts the git monitor if applicable.
// Note: Git monitor is per-repo and can only monitor a single directory.
// For multi-directory configs, git monitor uses the first directory only.
func (d *Daemon) startGitMonitor(ctx context.Context) error {
	watchDirs := d.Config.GetWatchDirs()
	if len(watchDirs) == 0 {
		return fmt.Errorf("no watch directories configured")
	}

	// Use the first watch directory for git monitoring
	// Git monitor is inherently per-repo, can't span multiple repos
	watchDir := watchDirs[0]
	if len(watchDirs) > 1 {
		log.Printf("Git monitor: using first directory only (%s), multi-directory git monitoring not supported", watchDir)
	}

	// Find git root from watch directory
	gitRoot, err := git.FindGitRoot(watchDir)
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

// startFileWatcher initializes and starts the file watcher if enabled.
// Note: File watcher currently supports single directory only.
// For multi-directory configs, uses the first directory.
func (d *Daemon) startFileWatcher(ctx context.Context) error {
	if !d.Config.Watcher.Enabled {
		log.Printf("File watcher disabled in config")
		return nil
	}

	watchDirs := d.Config.GetWatchDirs()
	if len(watchDirs) == 0 {
		return fmt.Errorf("no watch directories configured")
	}

	// Use the first watch directory for file watching
	watchDir := watchDirs[0]
	if len(watchDirs) > 1 {
		log.Printf("File watcher: using first directory only (%s), multi-directory file watching not yet implemented", watchDir)
	}

	// Create ignore matcher with default + config patterns
	ignore := watcher.NewIgnoreMatcherWithPatterns(d.Config.Watcher.IgnorePatterns)

	// Create watcher with config values
	fw, err := watcher.NewWatcher(watchDir, ignore, d.Config.Watcher.DebounceWindow)
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

// startTaskTrackers initializes and starts task trackers for all watch directories
func (d *Daemon) startTaskTrackers(ctx context.Context) error {
	if !d.Config.Task.Enabled {
		log.Printf("Task tracker disabled in config")
		return nil
	}

	// Initialize TaskSync for database synchronization
	d.taskSync = task.NewTaskSync(d.store)

	// Do initial sync for all directories
	for _, dir := range d.Config.GetWatchDirs() {
		count, err := d.taskSync.SyncDirectory(dir)
		if err != nil {
			log.Printf("Warning: initial task sync failed for %s: %v", dir, err)
		} else {
			log.Printf("Initial task sync: %d tasks from %s", count, dir)
		}
	}

	// Create a task tracker for each watch directory
	d.taskTrackers = make([]*task.TaskTracker, 0, len(d.Config.GetWatchDirs()))

	for _, dir := range d.Config.GetWatchDirs() {
		// Resolve task file path for this directory
		taskFilePath := d.Config.Task.FilePath
		if !filepath.IsAbs(taskFilePath) {
			taskFilePath = filepath.Join(dir, taskFilePath)
		}

		// Create task tracker
		tt, err := task.NewTracker(taskFilePath, d.Config.Task.PollInterval)
		if err != nil {
			log.Printf("Warning: failed to create task tracker for %s: %v", dir, err)
			continue
		}

		// Add daemon as listener (d implements ActivityListener)
		tt.AddListener(d)

		// Wire file change callback to trigger database sync
		tt.SetOnFileChanged(d.OnTaskFileChanged)

		// Start tracker
		if err := tt.Start(ctx); err != nil {
			log.Printf("Warning: failed to start task tracker for %s: %v", dir, err)
			continue
		}

		d.taskTrackers = append(d.taskTrackers, tt)
	}

	log.Printf("Task tracker monitoring: %d directories (poll every %v)", len(d.taskTrackers), d.Config.Task.PollInterval)
	return nil
}

// stopTaskTrackers stops all task trackers
func (d *Daemon) stopTaskTrackers() {
	for _, tt := range d.taskTrackers {
		if tt != nil {
			tt.Stop()
		}
	}
}

// OnTaskFileChanged is called when a TASKS.md file content changes.
// It triggers a database sync for the affected directory.
func (d *Daemon) OnTaskFileChanged(dir string) {
	log.Printf("TaskFileChanged: TASKS.md changed in %s", dir)
	if d.taskSync != nil {
		if count, err := d.taskSync.SyncDirectory(dir); err != nil {
			log.Printf("TaskFileChanged: Sync failed for %s: %v", dir, err)
		} else {
			log.Printf("TaskFileChanged: Synced %d tasks from %s", count, dir)
		}
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

// initBrainExecutor initializes the Python brain executor.
// Returns error if Python is not available or brain package not found.
func (d *Daemon) initBrainExecutor() error {
	executor := brain.NewExecutor()

	// Derive brain directory from executable path
	// During development: executable is in cmd/daemonflow/, brain is in project root
	// In production: brain package should be alongside executable or in a configured path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Walk up from executable to find brain/ directory
	brainDir := ""
	dir := filepath.Dir(exePath)
	for i := 0; i < 5; i++ { // Max 5 levels up
		if _, err := os.Stat(filepath.Join(dir, "brain", "__init__.py")); err == nil {
			brainDir = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Also try current working directory as fallback
	if brainDir == "" {
		if cwd, err := os.Getwd(); err == nil {
			if _, err := os.Stat(filepath.Join(cwd, "brain", "__init__.py")); err == nil {
				brainDir = cwd
			}
		}
	}

	if brainDir == "" {
		return fmt.Errorf("brain package not found (searched near executable and cwd)")
	}

	executor.SetBrainDir(brainDir)

	// Health check with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	version, err := executor.Ping(ctx)
	if err != nil {
		return fmt.Errorf("brain health check failed: %w", err)
	}

	d.brainExecutor = executor
	log.Printf("Python brain ready: version %s (dir: %s)", version, brainDir)
	return nil
}

// GetBrainExecutor returns the brain executor instance.
// May be nil if Python is not available or brain not initialized.
func (d *Daemon) GetBrainExecutor() *brain.Executor {
	return d.brainExecutor
}

// GetPendingTaskCount returns the count of incomplete tasks across all projects
func (d *Daemon) GetPendingTaskCount() int {
	if d.store == nil {
		return 0
	}

	count, err := d.store.CountIncompleteTasks()
	if err != nil {
		log.Printf("Warning: failed to count incomplete tasks: %v", err)
		return 0
	}
	return count
}

// GetGlobalTasks returns tasks from all projects, sorted by priority
func (d *Daemon) GetGlobalTasks(limit int, includeCompleted bool, projectPath string) ([]ipc.TaskData, int, error) {
	if d.store == nil {
		return nil, 0, fmt.Errorf("store not initialized")
	}

	var tasks []store.Task
	var err error

	if projectPath != "" {
		tasks, err = d.store.GetTasksByProject(projectPath)
	} else {
		tasks, err = d.store.GetTasksSortedByPriority()
	}

	if err != nil {
		return nil, 0, err
	}

	// Filter and convert
	result := make([]ipc.TaskData, 0, len(tasks))
	for _, t := range tasks {
		if !includeCompleted && t.Completed {
			continue
		}

		data := ipc.TaskData{
			ID:            t.ID,
			ProjectPath:   t.ProjectPath,
			ProjectName:   t.ProjectName,
			Text:          t.Text,
			Completed:     t.Completed,
			PriorityScore: t.PriorityScore,
			CreatedAt:     t.CreatedAt.Format(time.RFC3339),
		}
		if t.DueDate != nil {
			data.DueDate = t.DueDate.Format(time.RFC3339)
		}

		result = append(result, data)

		if len(result) >= limit {
			break
		}
	}

	return result, len(tasks), nil
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

	// Award XP based on activity type
	d.awardXPForActivity(act)
}

// awardXPForActivity awards experience points based on activity type.
// XP values:
//   - Commit: 10 XP
//   - Stage: 2 XP
//   - File save/change: 1 XP
//   - Task complete: 5 XP
func (d *Daemon) awardXPForActivity(act activity.Activity) {
	if d.store == nil {
		return // Store not initialized
	}

	var xpAmount int
	switch act.Type {
	case activity.GitCommit:
		xpAmount = 10
	case activity.GitStage:
		xpAmount = 2
	case activity.FileChange:
		xpAmount = 1
	case activity.TaskComplete:
		xpAmount = 5
	default:
		return // No XP for other activity types (e.g., branch switch)
	}

	if xpAmount > 0 {
		if _, err := d.store.AddExperience(xpAmount); err != nil {
			log.Printf("Warning: failed to award XP for %s: %v", act.Type, err)
		} else {
			log.Printf("Awarded %d XP for %s", xpAmount, act.Type)
		}
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

// GetPetLevelInfo returns the pet's level, experience, and XP needed for next level
func (d *Daemon) GetPetLevelInfo() (level, experience, experienceToNext int) {
	if d.store == nil {
		return 1, 0, 100 // Default: level 1, 0 XP, 100 to next
	}

	petState, err := d.store.GetPetState()
	if err != nil {
		log.Printf("Warning: failed to get pet state: %v", err)
		return 1, 0, 100
	}

	return petState.Level, petState.Experience, store.ExperienceToNextLevel(petState.Experience)
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

	// Sync graveyard stats to SQLite after resurrection
	d.syncGraveyardToStore()

	log.Printf("Pet resurrected! All earned time reset (total resurrections: %d)",
		d.graveyard.GetResurrectionCount())

	return &ipc.ResurrectResponse{
		Success:  true,
		Message:  "Pet resurrected! All earned time reset.",
		NewState: d.GetClockState(),
	}, nil
}

// syncGraveyardToStore syncs graveyard statistics (streaks, deaths, resurrections) to SQLite.
// Called on daemon start, after death, and after resurrection.
func (d *Daemon) syncGraveyardToStore() {
	if d.store == nil || d.graveyard == nil {
		return
	}

	streakInfo := d.graveyard.GetStreakInfo()
	totalDeaths := d.graveyard.GetDeathCount()
	totalResurrections := d.graveyard.GetResurrectionCount()

	err := d.store.SyncGraveyardStats(
		streakInfo.CurrentStreak,
		streakInfo.LongestStreak,
		totalDeaths,
		totalResurrections,
	)
	if err != nil {
		log.Printf("Warning: failed to sync graveyard stats to SQLite: %v", err)
	} else {
		log.Printf("Synced graveyard stats: streak=%d/%d, deaths=%d, resurrections=%d",
			streakInfo.CurrentStreak, streakInfo.LongestStreak, totalDeaths, totalResurrections)
	}
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
