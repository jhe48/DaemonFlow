package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the daemon configuration
type Config struct {
	// WatchDir is the directory to monitor for changes
	WatchDir string `yaml:"watch_dir"`

	// LogLevel controls logging verbosity (debug, info, warn, error)
	LogLevel string `yaml:"log_level"`

	// SocketPath is the IPC socket location for daemon communication
	SocketPath string `yaml:"socket_path"`

	// Git holds git monitoring configuration
	Git GitConfig `yaml:"git"`

	// Watcher holds file watcher configuration
	Watcher WatcherConfig `yaml:"watcher"`

	// Task holds task tracking configuration
	Task TaskConfig `yaml:"task"`

	// Earning holds Freedom Clock weights (placeholder for Phase 5)
	Earning EarningConfig `yaml:"earning"`
}

// GitConfig holds git monitoring settings
type GitConfig struct {
	// PollInterval is how often to check for git changes
	PollInterval time.Duration `yaml:"poll_interval"`
}

// WatcherConfig holds file watcher settings
type WatcherConfig struct {
	// Enabled controls whether file watching is active
	Enabled bool `yaml:"enabled"`

	// IgnorePatterns are additional patterns to ignore (beyond defaults)
	IgnorePatterns []string `yaml:"ignore_patterns"`

	// DebounceWindow is the time to wait before emitting batched file changes
	DebounceWindow time.Duration `yaml:"debounce_window"`
}

// EarningConfig holds Freedom Clock earning weights for each activity type
type EarningConfig struct {
	// CommitSeconds is seconds earned per git commit (default: 300 = 5 min)
	CommitSeconds int `yaml:"commit_seconds"`

	// StageSeconds is seconds earned per staging event (default: 60 = 1 min)
	StageSeconds int `yaml:"stage_seconds"`

	// FileChangeSeconds is seconds earned per file change batch (default: 30 = 30 sec)
	FileChangeSeconds int `yaml:"file_change_seconds"`

	// TaskCompleteSeconds is seconds earned per task completion (default: 180 = 3 min)
	TaskCompleteSeconds int `yaml:"task_complete_seconds"`
}

// TaskConfig holds task tracking settings
type TaskConfig struct {
	// Enabled controls whether task tracking is active
	Enabled bool `yaml:"enabled"`

	// FilePath is the path to the task file (relative to watch_dir)
	FilePath string `yaml:"file_path"`

	// PollInterval is how often to check for task changes
	PollInterval time.Duration `yaml:"poll_interval"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		homeDir = "."
	}

	cwd, _ := os.Getwd()
	if cwd == "" {
		cwd = "."
	}

	return &Config{
		WatchDir:   cwd,
		LogLevel:   "info",
		SocketPath: filepath.Join(homeDir, ".daemonflow", "daemonflow.sock"),
		Git: GitConfig{
			PollInterval: 5 * time.Second,
		},
		Watcher: WatcherConfig{
			Enabled:        true,
			IgnorePatterns: []string{}, // Use built-in defaults, allow user additions
			DebounceWindow: 500 * time.Millisecond,
		},
		Task: TaskConfig{
			Enabled:      true,
			FilePath:     "TASKS.md",
			PollInterval: 2 * time.Second,
		},
		Earning: EarningConfig{
			CommitSeconds:       300, // 5 minutes per commit
			StageSeconds:        60,  // 1 minute per staging event
			FileChangeSeconds:   30,  // 30 seconds per file change batch
			TaskCompleteSeconds: 180, // 3 minutes per task completion
		},
	}
}

// LoadConfig loads configuration from a YAML file
// If path is empty, it looks for ~/.daemonflow/config.yaml
// If the file doesn't exist, it returns default configuration
func LoadConfig(path string) (*Config, error) {
	config := DefaultConfig()

	// Determine config file path
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return config, nil // Return defaults if can't determine home
		}
		path = filepath.Join(homeDir, ".daemonflow", "config.yaml")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil // File doesn't exist, return defaults
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.LogLevel] {
		return fmt.Errorf("invalid log_level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	return nil
}
