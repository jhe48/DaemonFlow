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
	"github.com/jackyhe0402/daemonflow/internal/config"
	"github.com/jackyhe0402/daemonflow/internal/git"
	"github.com/jackyhe0402/daemonflow/internal/ipc"
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
			d.stopGitMonitor()
			return nil
		case <-d.shutdownChan:
			log.Printf("Received shutdown request via IPC, shutting down...")
			d.Running = false
			d.stopGitMonitor()
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
